import { useState } from "react";

import { AdminPanel } from "../admin/AdminPanel.jsx";
import { ProfilePanel } from "../auth/ProfilePanel.jsx";
import { LibraryPanel } from "../decks/LibraryPanel.jsx";
import { JournalPanel } from "../journal/JournalPanel.jsx";
import { LessonsPanel } from "../lessons/LessonsPanel.jsx";
import { ReviewPanel } from "../review/ReviewPanel.jsx";
import { ReviewQueue } from "../review/ReviewQueue.jsx";
import { SearchPanel } from "../search/SearchPanel.jsx";
import { InsightsPanel } from "../stats/InsightsPanel.jsx";
import { MetricCard } from "./components/MetricCard.jsx";
import { useStudyDashboard } from "./hooks/useStudyDashboard.js";
import { Sidebar } from "../../shared/ui/Sidebar.jsx";

const titles = {
  study: ["Aujourd'hui", "Revision du jour"],
  library: ["Collection", "Bibliotheque"],
  lessons: ["Parcours guidé", "Leçons de coréen"],
  journal: ["Pratique libre", "Journal en coreen"],
  insights: ["Regularite", "Progression"],
  search: ["Retrouver", "Recherche globale"],
  profile: ["Compte", "Profil et administration"]
};

export function DashboardPage({ authToken, currentUser, onLogout, onUpdateProfile }) {
  const [view, setView] = useState("study");
  const dashboard = useStudyDashboard(authToken);
  const [eyebrow, title] = titles[view];

  return (
    <main className="shell">
      <Sidebar activeView={view} apiOnline={dashboard.apiOnline} currentUser={currentUser} onLogout={onLogout} onNavigate={setView} />
      <section className="workspace">
        <header className="topbar"><div><p className="eyebrow">{eyebrow}</p><h1>{title}</h1></div>{view === "study" && <button className="primary-button" onClick={dashboard.reload} type="button">Actualiser</button>}</header>
        {dashboard.error && <div className="error-banner"><strong>Connexion ou operation impossible</strong><span>{dashboard.error}</span><button onClick={dashboard.reload} type="button">Reessayer</button></div>}

        {view === "study" && <>
          <section className="metrics" aria-label="Statistiques du jour"><MetricCard label="A reviser" value={dashboard.stats.dueCards} /><MetricCard label="Nouvelles" value={dashboard.stats.newCards} /><MetricCard label="Difficiles" value={dashboard.stats.difficultCards} /><MetricCard label="Serie" value={`${dashboard.stats.currentStreak} j`} /></section>
          <section className="study-layout"><ReviewPanel activeIndex={dashboard.activeIndex} card={dashboard.activeCard} isLoading={dashboard.isLoading || dashboard.isMutating} onAnswer={dashboard.answerCard} onCheck={dashboard.checkAnswer} totalCards={dashboard.dueCards.length} /><ReviewQueue activeIndex={dashboard.activeIndex} cards={dashboard.dueCards} isLoading={dashboard.isLoading || dashboard.isMutating} onSelect={dashboard.selectCard} /></section>
        </>}
        {view === "library" && <LibraryPanel cards={dashboard.cards} decks={dashboard.decks} isMutating={dashboard.isMutating} runMutation={dashboard.runMutation} token={authToken} />}
        {view === "lessons" && <LessonsPanel isMutating={dashboard.isMutating} lessons={dashboard.lessons} runMutation={dashboard.runMutation} token={authToken} />}
        {view === "journal" && <JournalPanel entries={dashboard.journal} isMutating={dashboard.isMutating} runMutation={dashboard.runMutation} token={authToken} />}
        {view === "insights" && <InsightsPanel difficultCards={dashboard.difficultCards} stats={dashboard.stats} />}
        {view === "search" && <SearchPanel token={authToken} />}
        {view === "profile" && <div className="content-stack"><ProfilePanel currentUser={currentUser} onUpdateProfile={onUpdateProfile} />{currentUser?.isAdmin && <AdminPanel isMutating={dashboard.isMutating} runMutation={dashboard.runMutation} token={authToken} />}</div>}
      </section>
    </main>
  );
}
