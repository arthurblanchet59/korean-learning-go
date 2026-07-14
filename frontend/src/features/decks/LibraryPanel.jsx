import { useEffect, useMemo, useState } from "react";

import {
  bulkDeleteCards,
  bulkDeleteDecks,
  bulkUpdateCards,
  bulkUpdateDecks,
  createCard,
  createDeck,
  deleteCard,
  deleteDeck,
  exportCardsCSV,
  importCardsCSV,
  updateCard,
  updateDeck
} from "../../shared/api/studyApi.js";
import { CardEditorModal } from "./CardEditorModal.jsx";

const emptyDeck = { name: "", description: "" };
const emptyCard = { deckId: "", kind: "vocabulary", korean: "", translation: "", romanization: "", exampleKorean: "", exampleTranslation: "", tags: "" };
const kindLabels = { vocabulary: "Vocabulaire", phrase: "Phrase", hangul: "Hangeul" };
const cardsPerPage = 15;

export function LibraryPanel({ cards, decks, isMutating, runMutation, token }) {
  const [deckForm, setDeckForm] = useState(emptyDeck);
  const [deckEditing, setDeckEditing] = useState(null);
  const [cardForm, setCardForm] = useState(emptyCard);
  const [cardEditing, setCardEditing] = useState(null);
  const [cardModalOpen, setCardModalOpen] = useState(false);
  const [cardQuery, setCardQuery] = useState("");
  const [deckFilter, setDeckFilter] = useState("all");
  const [kindFilter, setKindFilter] = useState("all");
  const [cardSort, setCardSort] = useState("newest");
  const [cardPage, setCardPage] = useState(1);
  const [selectedCards, setSelectedCards] = useState([]);
  const [selectedDecks, setSelectedDecks] = useState([]);
  const [bulkDeckDescription, setBulkDeckDescription] = useState("");

  const decksById = useMemo(() => new Map(decks.map((deck) => [deck.id, deck])), [decks]);
  const visibleCards = useMemo(() => {
    const query = cardQuery.trim().toLocaleLowerCase("fr");
    return cards
      .filter((card) => deckFilter === "all" || card.deckId === deckFilter)
      .filter((card) => kindFilter === "all" || card.kind === kindFilter)
      .filter((card) => {
        if (!query) return true;
        return [
          card.korean,
          card.translation,
          card.romanization,
          card.exampleKorean,
          card.exampleTranslation,
          ...(card.tags ?? []),
          decksById.get(card.deckId)?.name
        ].filter(Boolean).join(" ").toLocaleLowerCase("fr").includes(query);
      })
      .sort((left, right) => compareCards(left, right, cardSort, decksById));
  }, [cardQuery, cardSort, cards, deckFilter, decksById, kindFilter]);
  const pageCount = Math.max(1, Math.ceil(visibleCards.length / cardsPerPage));
  const paginatedCards = visibleCards.slice((cardPage - 1) * cardsPerPage, cardPage * cardsPerPage);

  useEffect(() => {
    setCardPage(1);
  }, [cardQuery, cardSort, deckFilter, kindFilter]);

  useEffect(() => {
    if (cardPage > pageCount) setCardPage(pageCount);
  }, [cardPage, pageCount]);

  async function submitDeck(event) {
    event.preventDefault();
    const result = await runMutation(() => deckEditing ? updateDeck(deckEditing, deckForm, token) : createDeck(deckForm, token));
    if (result.ok) {
      setDeckForm(emptyDeck);
      setDeckEditing(null);
    }
  }

  async function submitCard(event) {
    event.preventDefault();
    const payload = {
      ...cardForm,
      deckId: cardForm.deckId || decks[0]?.id || "",
      tags: cardForm.tags.split(",").map((tag) => tag.trim()).filter(Boolean)
    };
    const result = await runMutation(() => cardEditing ? updateCard(cardEditing, payload, token) : createCard(payload, token));
    if (result.ok) closeCardModal();
  }

  function editDeck(deck) {
    setDeckEditing(deck.id);
    setDeckForm({ name: deck.name, description: deck.description });
  }

  function openCreateCard() {
    setCardEditing(null);
    setCardForm({ ...emptyCard, deckId: deckFilter !== "all" ? deckFilter : decks[0]?.id ?? "" });
    setCardModalOpen(true);
  }

  function openEditCard(card) {
    setCardEditing(card.id);
    setCardForm({
      deckId: card.deckId,
      kind: card.kind,
      korean: card.korean,
      translation: card.translation,
      romanization: card.romanization ?? "",
      exampleKorean: card.exampleKorean ?? "",
      exampleTranslation: card.exampleTranslation ?? "",
      tags: (card.tags ?? []).join(", ")
    });
    setCardModalOpen(true);
  }

  function closeCardModal() {
    setCardModalOpen(false);
    setCardEditing(null);
    setCardForm(emptyCard);
  }

  function toggleSelection(setter, current, id) {
    setter(current.includes(id) ? current.filter((value) => value !== id) : [...current, id]);
  }

  async function handleCSV(event) {
    const file = event.target.files?.[0];
    const targetDeck = deckFilter !== "all" ? deckFilter : decks[0]?.id;
    if (!file || !targetDeck) return;
    const content = await file.text();
    await runMutation(() => importCardsCSV(targetDeck, content, token));
    event.target.value = "";
  }

  async function exportCSV() {
    const result = await exportCardsCSV(token);
    if (!result.ok) return;
    const url = URL.createObjectURL(new Blob([result.data], { type: "text/csv;charset=utf-8" }));
    const link = document.createElement("a");
    link.href = url;
    link.download = "korean-cards.csv";
    link.click();
    URL.revokeObjectURL(url);
  }

  return (
    <fieldset aria-busy={isMutating} className="mutation-surface" disabled={isMutating}>
      <div className="content-stack">
        <section className="management-section">
          <div className="section-heading"><div><p className="eyebrow">Organisation</p><h2>Decks</h2></div><strong>{decks.length}</strong></div>
          <form className="inline-form" onSubmit={submitDeck}>
            <input aria-label="Nom du deck" onChange={(event) => setDeckForm({ ...deckForm, name: event.target.value })} placeholder="Nom du deck" required value={deckForm.name} />
            <input aria-label="Description du deck" onChange={(event) => setDeckForm({ ...deckForm, description: event.target.value })} placeholder="Description" value={deckForm.description} />
            <button className="primary-button" type="submit">{deckEditing ? "Modifier" : "Ajouter"}</button>
            {deckEditing && <button className="secondary-button" onClick={() => { setDeckEditing(null); setDeckForm(emptyDeck); }} type="button">Annuler</button>}
          </form>
          {selectedDecks.length > 0 && <div className="bulk-bar"><span>{selectedDecks.length} sélectionné(s)</span><input onChange={(event) => setBulkDeckDescription(event.target.value)} placeholder="Description commune" value={bulkDeckDescription} /><button className="secondary-button" disabled={!bulkDeckDescription} onClick={() => runMutation(() => bulkUpdateDecks(selectedDecks, { description: bulkDeckDescription }, token)).then(() => setBulkDeckDescription(""))} type="button">Appliquer</button><button className="danger-button" onClick={() => window.confirm("Supprimer les decks sélectionnés et toutes leurs cartes ?") && runMutation(() => bulkDeleteDecks(selectedDecks, token)).then(() => setSelectedDecks([]))} type="button">Supprimer la sélection</button></div>}
          <div className="data-list">
            {decks.map((deck) => <article className="data-row" key={deck.id}>
              <input aria-label={`Sélectionner ${deck.name}`} checked={selectedDecks.includes(deck.id)} onChange={() => toggleSelection(setSelectedDecks, selectedDecks, deck.id)} type="checkbox" />
              <div><strong>{deck.name}</strong><span>{deck.description || "Sans description"}</span></div>
              <button className="secondary-button" onClick={() => editDeck(deck)} type="button">Modifier</button>
              <button className="danger-button" onClick={() => window.confirm(`Supprimer « ${deck.name} » et toutes ses cartes ?`) && runMutation(() => deleteDeck(deck.id, token))} type="button">Supprimer</button>
            </article>)}
          </div>
        </section>

        <section className="management-section card-library">
          <div className="section-heading card-library-heading">
            <div><p className="eyebrow">Contenu</p><h2>Cartes</h2></div>
            <div className="button-row">
              <label className="secondary-button file-button">Importer CSV<input accept=".csv,text/csv" onChange={handleCSV} type="file" /></label>
              <button className="secondary-button" onClick={exportCSV} type="button">Exporter CSV</button>
              <button className="primary-button" disabled={decks.length === 0} onClick={openCreateCard} type="button">Ajouter une carte</button>
            </div>
          </div>

          <div className="card-library-toolbar" aria-label="Filtres des cartes">
            <label className="card-search-field">
              <span>Rechercher</span>
              <input onChange={(event) => setCardQuery(event.target.value)} placeholder="Coréen, traduction, romanisation ou tag" type="search" value={cardQuery} />
            </label>
            <label><span>Deck</span><select onChange={(event) => setDeckFilter(event.target.value)} value={deckFilter}><option value="all">Tous les decks</option>{decks.map((deck) => <option key={deck.id} value={deck.id}>{deck.name}</option>)}</select></label>
            <label><span>Type</span><select onChange={(event) => setKindFilter(event.target.value)} value={kindFilter}><option value="all">Tous les types</option><option value="vocabulary">Vocabulaire</option><option value="phrase">Phrase</option><option value="hangul">Hangeul</option></select></label>
            <label><span>Trier par</span><select onChange={(event) => setCardSort(event.target.value)} value={cardSort}><option value="newest">Plus récentes</option><option value="korean">Coréen A-Z</option><option value="translation">Traduction A-Z</option><option value="deck">Deck</option></select></label>
          </div>

          <div className="card-results-summary" role="status"><strong>{visibleCards.length}</strong> carte(s) affichée(s) sur {cards.length}</div>
          {selectedCards.length > 0 && <div className="bulk-bar"><span>{selectedCards.length} sélectionnée(s)</span><select onChange={(event) => event.target.value && runMutation(() => bulkUpdateCards(selectedCards, { deckId: event.target.value }, token)).then((result) => result.ok && setSelectedCards([]))} defaultValue=""><option value="">Déplacer vers...</option>{decks.map((deck) => <option key={deck.id} value={deck.id}>{deck.name}</option>)}</select><button className="danger-button" onClick={() => window.confirm("Supprimer les cartes sélectionnées ?") && runMutation(() => bulkDeleteCards(selectedCards, token)).then((result) => result.ok && setSelectedCards([]))} type="button">Supprimer</button></div>}

          <div className="card-management-list">
            {visibleCards.length === 0 && <div className="library-empty-state"><strong>Aucune carte trouvée</strong><span>Modifie les filtres ou ajoute une nouvelle carte.</span></div>}
            {paginatedCards.map((card) => {
              const deck = decksById.get(card.deckId);
              return <article aria-label={`Carte ${card.korean}`} className="managed-card-row" key={card.id}>
                <input aria-label={`Sélectionner ${card.korean}`} checked={selectedCards.includes(card.id)} onChange={() => toggleSelection(setSelectedCards, selectedCards, card.id)} type="checkbox" />
                <div className="managed-card-content">
                  <div><strong className="korean-text">{card.korean}</strong>{card.romanization && <span>{card.romanization}</span>}</div>
                  <p>{card.translation}</p>
                  <div className="managed-card-meta"><span>{deck?.name ?? "Deck inconnu"}</span><span>{kindLabels[card.kind] ?? card.kind}</span>{card.tags?.map((tag) => <span key={tag}>#{tag}</span>)}</div>
                </div>
                <div className="managed-card-actions"><button className="secondary-button" onClick={() => openEditCard(card)} type="button">Modifier</button><button className="danger-button" onClick={() => window.confirm(`Supprimer la carte « ${card.korean} » ?`) && runMutation(() => deleteCard(card.id, token))} type="button">Supprimer</button></div>
              </article>;
            })}
          </div>
          {pageCount > 1 && <nav aria-label="Pagination des cartes" className="card-pagination"><button className="secondary-button" disabled={cardPage === 1} onClick={() => setCardPage((page) => page - 1)} type="button">Précédent</button><span>Page {cardPage} sur {pageCount}</span><button className="secondary-button" disabled={cardPage === pageCount} onClick={() => setCardPage((page) => page + 1)} type="button">Suivant</button></nav>}
        </section>
      </div>

      <CardEditorModal
        decks={decks}
        form={cardForm}
        isEditing={Boolean(cardEditing)}
        onChange={setCardForm}
        onClose={closeCardModal}
        onSubmit={submitCard}
        open={cardModalOpen}
      />
    </fieldset>
  );
}

function compareCards(left, right, sort, decksById) {
  if (sort === "korean") return left.korean.localeCompare(right.korean, "ko");
  if (sort === "translation") return left.translation.localeCompare(right.translation, "fr");
  if (sort === "deck") {
    const deckComparison = (decksById.get(left.deckId)?.name ?? "").localeCompare(decksById.get(right.deckId)?.name ?? "", "fr");
    return deckComparison || left.korean.localeCompare(right.korean, "ko");
  }
  return new Date(right.createdAt).getTime() - new Date(left.createdAt).getTime();
}
