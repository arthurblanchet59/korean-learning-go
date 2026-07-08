import { DeckList } from "../decks/DeckList.jsx";
import { ReviewPanel } from "../review/ReviewPanel.jsx";
import { ReviewQueue } from "../review/ReviewQueue.jsx";
import { MetricCard } from "./components/MetricCard.jsx";
import { useStudyDashboard } from "./hooks/useStudyDashboard.js";
import { Sidebar } from "../../shared/ui/Sidebar.jsx";

export function DashboardPage() {
  const {
    apiOnline,
    cards,
    decks,
    stats,
    activeCard,
    activeIndex,
    isLoading,
    answerCard,
    selectCard,
    restartSession
  } = useStudyDashboard();

  return (
    <main className="shell">
      <Sidebar apiOnline={apiOnline} />

      <section className="workspace">
        <header className="topbar">
          <div>
            <p className="eyebrow">Aujourd'hui</p>
            <h1>Revision coreen</h1>
          </div>
          <button className="primary-button" type="button" onClick={restartSession}>
            Commencer
          </button>
        </header>

        <section className="metrics" aria-label="Statistiques du jour">
          <MetricCard label="A reviser" value={stats.dueCards} />
          <MetricCard label="Nouvelles" value={stats.newCards} />
          <MetricCard label="Difficiles" value={stats.difficultCards} />
        </section>

        <section className="study-layout">
          <ReviewPanel
            card={activeCard}
            activeIndex={activeIndex}
            totalCards={cards.length}
            isLoading={isLoading}
            onAnswer={answerCard}
          />
          <ReviewQueue cards={cards} activeIndex={activeIndex} onSelect={selectCard} />
        </section>

        <DeckList decks={decks} />
      </section>
    </main>
  );
}
