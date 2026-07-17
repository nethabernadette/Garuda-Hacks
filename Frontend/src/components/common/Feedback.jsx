export function Feedback({ status, error, empty, children }) {
  if (status === 'loading') return <div className="state-card">Memuat data...</div>;
  if (status === 'error') return <div className="state-card error">Gagal memuat data. {error}</div>;
  if (empty) return <div className="state-card">Data belum tersedia.</div>;
  return children;
}
