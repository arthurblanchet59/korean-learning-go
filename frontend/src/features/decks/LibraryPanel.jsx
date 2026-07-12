import { useState } from "react";

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

const emptyDeck = { name: "", description: "" };
const emptyCard = { deckId: "", kind: "vocabulary", korean: "", translation: "", romanization: "", exampleKorean: "", exampleTranslation: "", tags: "" };

export function LibraryPanel({ cards, decks, isMutating, runMutation, token }) {
  const [deckForm, setDeckForm] = useState(emptyDeck);
  const [deckEditing, setDeckEditing] = useState(null);
  const [cardForm, setCardForm] = useState(emptyCard);
  const [cardEditing, setCardEditing] = useState(null);
  const [selectedCards, setSelectedCards] = useState([]);
  const [selectedDecks, setSelectedDecks] = useState([]);
  const [bulkDeckDescription, setBulkDeckDescription] = useState("");

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
    const payload = { ...cardForm, deckId: cardForm.deckId || decks[0]?.id || "", tags: cardForm.tags.split(",").map((tag) => tag.trim()).filter(Boolean) };
    const result = await runMutation(() => cardEditing ? updateCard(cardEditing, payload, token) : createCard(payload, token));
    if (result.ok) {
      setCardForm({ ...emptyCard, deckId: decks[0]?.id ?? "" });
      setCardEditing(null);
    }
  }

  function editDeck(deck) {
    setDeckEditing(deck.id);
    setDeckForm({ name: deck.name, description: deck.description });
  }

  function editCard(card) {
    setCardEditing(card.id);
    setCardForm({ ...card, tags: (card.tags ?? []).join(", ") });
  }

  function toggleSelection(setter, current, id) {
    setter(current.includes(id) ? current.filter((value) => value !== id) : [...current, id]);
  }

  async function handleCSV(event) {
    const file = event.target.files?.[0];
    const targetDeck = cardForm.deckId || decks[0]?.id;
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
        {selectedDecks.length > 0 && <div className="bulk-bar"><span>{selectedDecks.length} selectionne(s)</span><input onChange={(event) => setBulkDeckDescription(event.target.value)} placeholder="Description commune" value={bulkDeckDescription} /><button className="secondary-button" disabled={!bulkDeckDescription} onClick={() => runMutation(() => bulkUpdateDecks(selectedDecks, { description: bulkDeckDescription }, token)).then(() => setBulkDeckDescription(""))} type="button">Appliquer</button><button className="danger-button" onClick={() => window.confirm("Supprimer les decks sélectionnés et toutes leurs cartes ?") && runMutation(() => bulkDeleteDecks(selectedDecks, token)).then(() => setSelectedDecks([]))} type="button">Supprimer la selection</button></div>}
        <div className="data-list">
          {decks.map((deck) => <article className="data-row" key={deck.id}>
            <input aria-label={`Selectionner ${deck.name}`} checked={selectedDecks.includes(deck.id)} onChange={() => toggleSelection(setSelectedDecks, selectedDecks, deck.id)} type="checkbox" />
            <div><strong>{deck.name}</strong><span>{deck.description || "Sans description"}</span></div>
            <button className="secondary-button" onClick={() => editDeck(deck)} type="button">Modifier</button>
            <button className="danger-button" onClick={() => window.confirm(`Supprimer « ${deck.name} » et toutes ses cartes ?`) && runMutation(() => deleteDeck(deck.id, token))} type="button">Supprimer</button>
          </article>)}
        </div>
      </section>

      <section className="management-section">
        <div className="section-heading"><div><p className="eyebrow">Contenu</p><h2>Cartes</h2></div><div className="button-row"><label className="secondary-button file-button">Importer CSV<input accept=".csv,text/csv" onChange={handleCSV} type="file" /></label><button className="secondary-button" onClick={exportCSV} type="button">Exporter CSV</button></div></div>
        <form className="card-form" onSubmit={submitCard}>
          <select onChange={(event) => setCardForm({ ...cardForm, deckId: event.target.value })} required value={cardForm.deckId || decks[0]?.id || ""}><option value="">Choisir un deck</option>{decks.map((deck) => <option key={deck.id} value={deck.id}>{deck.name}</option>)}</select>
          <select onChange={(event) => setCardForm({ ...cardForm, kind: event.target.value })} value={cardForm.kind}><option value="vocabulary">Vocabulaire</option><option value="phrase">Phrase</option><option value="hangul">Hangeul</option></select>
          <input onChange={(event) => setCardForm({ ...cardForm, korean: event.target.value })} placeholder="Coreen" required value={cardForm.korean} />
          <input onChange={(event) => setCardForm({ ...cardForm, translation: event.target.value })} placeholder="Traduction" required value={cardForm.translation} />
          <input onChange={(event) => setCardForm({ ...cardForm, romanization: event.target.value })} placeholder="Romanisation" value={cardForm.romanization} />
          <input onChange={(event) => setCardForm({ ...cardForm, exampleKorean: event.target.value })} placeholder="Exemple coreen" value={cardForm.exampleKorean} />
          <input onChange={(event) => setCardForm({ ...cardForm, exampleTranslation: event.target.value })} placeholder="Traduction de l'exemple" value={cardForm.exampleTranslation} />
          <input onChange={(event) => setCardForm({ ...cardForm, tags: event.target.value })} placeholder="Tags separes par des virgules" value={cardForm.tags} />
          <button className="primary-button" type="submit">{cardEditing ? "Modifier" : "Ajouter la carte"}</button>
          {cardEditing && <button className="secondary-button" onClick={() => { setCardEditing(null); setCardForm(emptyCard); }} type="button">Annuler</button>}
        </form>
        {selectedCards.length > 0 && <div className="bulk-bar"><span>{selectedCards.length} selectionnee(s)</span><select onChange={(event) => event.target.value && runMutation(() => bulkUpdateCards(selectedCards, { deckId: event.target.value }, token))} defaultValue=""><option value="">Deplacer vers...</option>{decks.map((deck) => <option key={deck.id} value={deck.id}>{deck.name}</option>)}</select><button className="danger-button" onClick={() => window.confirm("Supprimer les cartes sélectionnées ?") && runMutation(() => bulkDeleteCards(selectedCards, token)).then(() => setSelectedCards([]))} type="button">Supprimer</button></div>}
        <div className="data-list">
          {cards.map((card) => <article className="data-row card-row" key={card.id}>
            <input aria-label={`Selectionner ${card.korean}`} checked={selectedCards.includes(card.id)} onChange={() => toggleSelection(setSelectedCards, selectedCards, card.id)} type="checkbox" />
            <div><strong className="korean-text">{card.korean}</strong><span>{card.translation} · {card.romanization}</span></div>
            <span className="tag">{card.kind}</span>
            <button className="secondary-button" onClick={() => editCard(card)} type="button">Modifier</button>
            <button className="danger-button" onClick={() => window.confirm(`Supprimer la carte « ${card.korean} » ?`) && runMutation(() => deleteCard(card.id, token))} type="button">Supprimer</button>
          </article>)}
        </div>
      </section>
    </div>
    </fieldset>
  );
}
