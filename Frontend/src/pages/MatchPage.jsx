import { Link } from 'react-router-dom';
import { PhoneShell } from '../components/layout/PhoneShell.jsx';
import { Button } from '../components/common/Button.jsx';
import { demoMatches } from '../data/demoData.js';

export function MatchPage() {
  return (
    <PhoneShell title="AI Matching" gradient backTo="/home">
      <div className="body">
        <div className="ali-say pad">
          <span className="ali-mini">🌱</span>
          <div className="bubble"><b>Ali:</b> Aku menemukan mitra dengan kecocokan tinggi berdasarkan kategori, lokasi, kapasitas, dan budget.</div>
        </div>
        {demoMatches.map((match) => (
          <article className="card match-big" key={match.id}>
            <div className="head">
              <span className="logo">🌶️</span>
              <div className="nm">
                <b>{match.company}</b>
                <small>{match.city}</small>
              </div>
              <div className="score-ring"><b>{match.score}%</b></div>
            </div>
            <div className="kv">
              <div><small>Produk</small><b>{match.product}</b></div>
              <div><small>Kapasitas</small><b>{match.capacity}</b></div>
              <div><small>Harga</small><b>{match.price}</b></div>
              <div><small>Kategori</small><b>{match.category}</b></div>
            </div>
            <div className="why">
              <b>Kenapa cocok?</b>
              <p>{match.reason}</p>
            </div>
            <div className="mb-actions">
              <Link className="btn btn-ghost" to="/home">Lewati</Link>
              <Link className="btn btn-teal" to="/chat">Saya Berminat</Link>
            </div>
          </article>
        ))}
      </div>
    </PhoneShell>
  );
}
