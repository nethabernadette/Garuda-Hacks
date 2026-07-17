export function Button({ children, variant = 'navy', loading = false, ...props }) {
  return (
    <button className={`btn btn-${variant}`} disabled={loading || props.disabled} {...props}>
      {loading ? 'Memproses...' : children}
    </button>
  );
}
