export function MetricCard({ label, value }) {
  return (
    <article>
      <span>{label}</span>
      <strong>{value ?? "--"}</strong>
    </article>
  );
}
