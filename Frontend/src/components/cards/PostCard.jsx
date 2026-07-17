export function PostCard({ item }) {
  const postType = item?.post_type || item?.postType || 'supply';
  const isSupply = postType === 'supply';
  const price = isSupply
    ? range(item?.price_min, item?.price_max)
    : range(item?.budget_min, item?.budget_max);

  return (
    <article className="card mcard">
      <div className="top">
        <span className="logo">{isSupply ? '🌾' : '🛒'}</span>
        <div>
          <b>{item?.product_name || item?.productName || 'Produk pangan'}</b>
          <small>{item?.category || 'Kategori'} · {item?.location || 'Lokasi belum tersedia'}</small>
        </div>
        <span className="score-mini">{item?.relevance_score || 0}%</span>
      </div>
      <div className="kv mini">
        <div><small>{isSupply ? 'Stok' : 'Butuh'}</small><b>{item?.quantity || '-'} {item?.unit || ''}</b></div>
        <div><small>{isSupply ? 'Harga' : 'Budget'}</small><b>{price}</b></div>
      </div>
      <div className="foot">
        <span className="chip">{postType}</span>
        <span className="chip">{item?.status || 'active'}</span>
      </div>
    </article>
  );
}

function range(min, max) {
  if (!min && !max) return 'Belum tersedia';
  if (min && max) return `Rp${Number(min).toLocaleString('id-ID')} - Rp${Number(max).toLocaleString('id-ID')}`;
  return `Rp${Number(min || max).toLocaleString('id-ID')}`;
}
