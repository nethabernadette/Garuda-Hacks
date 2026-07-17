import { Link } from 'react-router-dom';
import { PhoneShell } from '../components/layout/PhoneShell.jsx';
import { Feedback } from '../components/common/Feedback.jsx';
import { PostCard } from '../components/cards/PostCard.jsx';
import { api } from '../services/api.js';
import { useAsync } from '../hooks/useAsync.js';
import { useAuth } from '../context/AuthContext.jsx';

export function HomePage() {
  const auth = useAuth();
  const role = auth.profile?.role || 'BUYER';
  const type = role === 'BUYER' ? 'supply' : 'demand';
  const feed = useAsync(() => api.feed({ type, limit: 5 }), [type]);
  const items = feed.data?.items || [];

  return (
    <PhoneShell title="" gradient>
      <div className="body">
        <div className="home-hero">
          <div className="home-head">
            <span className="avatar">👩🏻‍🌾</span>
            <div className="who">
              <small>Selamat datang</small>
              <b>{auth.profile?.company_name || 'UD Berkah Pangan'}</b>
            </div>
            <Link className="bell" to="/notifications">🔔<i /></Link>
          </div>
          <Link className="cta" to="/post">
            <span className="ali-mini">🌱</span>
            <span>
              <b>{role === 'BUYER' ? 'Posting kebutuhan bahan pangan' : 'Tawarkan supply terbaru'}</b>
              <p>Ali akan bantu mencari mitra paling relevan dari feed dan kategori bisnis.</p>
            </span>
            <span className="go">›</span>
          </Link>
        </div>
        <div className="pad">
          <div className="stats">
            <div className="card stat"><small>Match Baru</small><div className="num">12</div><span className="unit">mitra</span></div>
            <div className="card stat"><small>Negosiasi</small><div className="num">3</div><span className="unit">aktif</span></div>
          </div>
          <div className="sec-h">
            <b>{role === 'BUYER' ? 'Supply cocok untukmu' : 'Demand terbuka'}</b>
            <Link to="/match">Lihat</Link>
          </div>
          <Feedback status={feed.status} error={feed.error} empty={feed.status === 'success' && items.length === 0}>
            {items.map((item) => <PostCard key={item.id} item={item} />)}
          </Feedback>
        </div>
      </div>
    </PhoneShell>
  );
}
