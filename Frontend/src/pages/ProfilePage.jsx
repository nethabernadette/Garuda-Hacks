import { Link } from 'react-router-dom';
import { PhoneShell } from '../components/layout/PhoneShell.jsx';
import { Button } from '../components/common/Button.jsx';
import { useAuth } from '../context/AuthContext.jsx';

export function ProfilePage() {
  const auth = useAuth();
  const profile = auth.profile || {};

  return (
    <PhoneShell title="">
      <div className="body profile-body">
        <div className="prof-hero">
          <span className="pf">👩🏻‍🌾</span>
          <h3>{profile.company_name || 'UD Berkah Pangan'}</h3>
          <p>{profile.role || 'BUYER'} · {profile.city || 'Bandung, Jawa Barat'}</p>
          <span className="badge-nib">✓ Profil Bisnis</span>
        </div>
        <div className="card prof-stats">
          <div><b>7</b><small>Kesepakatan</small></div>
          <div><b>4.9</b><small>Rating Mitra</small></div>
          <div><b>26</b><small>Koneksi</small></div>
        </div>
        <div className="pad">
          <div className="sec-h"><b>Komoditas</b><Link to="/profile-setup">Edit</Link></div>
          <div className="chip-row">
            {(profile.product_category || 'Cabai Merah,Bawang Merah,Tomat').split(',').map((item) => <span key={item} className="chip">{item.trim()}</span>)}
          </div>
          <div className="sec-h"><b>Pengaturan</b></div>
          <Link className="card nrow" to="/profile-setup"><span className="em">🏢</span><span><b>Edit Profil Bisnis</b><p>Nama usaha, kontak, kapasitas, lokasi</p></span></Link>
          <Link className="card nrow" to="/nib"><span className="em">🛡️</span><span><b>Verifikasi NIB</b><p>Perkuat trust antar mitra</p></span></Link>
          <Link className="card nrow" to="/notifications"><span className="em">🔔</span><span><b>Notifikasi</b><p>Match, penawaran, kesepakatan</p></span></Link>
          <Button variant="ghost" onClick={auth.logout}>Keluar</Button>
        </div>
      </div>
    </PhoneShell>
  );
}
