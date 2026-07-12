export function InsightsPanel({ difficultCards, stats }) {
  const maxReviews = Math.max(1, ...(stats.reviewHistory ?? []).map((day) => day.reviews));
  return (
    <div className="content-stack">
      <section className="metrics metrics-wide">
        <article><span>Revues aujourd'hui</span><strong>{stats.reviewsToday}</strong></article>
        <article><span>Precision</span><strong>{Math.round(stats.accuracyPercent)}%</strong></article>
        <article><span>Serie actuelle</span><strong>{stats.currentStreak} j</strong></article>
        <article><span>Maitrisees</span><strong>{stats.masteredCards}</strong></article>
      </section>
      <section className="management-section"><div className="section-heading"><div><p className="eyebrow">90 derniers jours</p><h2>Rythme de revision</h2></div><span>Record : {stats.longestStreak} jours</span></div><div className="history-chart">{(stats.reviewHistory ?? []).map((day) => <div className="history-bar" key={day.date} title={`${day.date}: ${day.reviews} revisions`} style={{ height: `${Math.max(8, day.reviews / maxReviews * 100)}%` }} />)}</div></section>
      <section className="management-section"><div className="section-heading"><div><p className="eyebrow">A renforcer</p><h2>Cartes difficiles</h2></div><strong>{difficultCards.length}</strong></div><div className="data-list">{difficultCards.map((card) => <article className="data-row" key={card.id}><div><strong className="korean-text">{card.korean}</strong><span>{card.translation}</span></div><span>{card.reviewState.lapseCount} oublis</span></article>)}</div></section>
    </div>
  );
}
