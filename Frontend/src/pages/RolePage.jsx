import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { PhoneShell } from '../components/layout/PhoneShell.jsx';
import { Button } from '../components/common/Button.jsx';
import { Field } from '../components/forms/Field.jsx';
import { useAuth } from '../context/AuthContext.jsx';

export function RolePage() {
  const navigate = useNavigate();
  const auth = useAuth();
  const [role, setRole] = useState('BUYER');
  const [email, setEmail] = useState('buyer@example.com');
  const [password, setPassword] = useState('password123');
  const [mode, setMode] = useState('register');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  async function submit(event) {
    event.preventDefault();
    setLoading(true);
    setError('');
    try {
      if (mode === 'register') {
        await auth.register(role, email, password);
      } else {
        await auth.login(email, password);
      }
      navigate('/profile-setup');
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }

  return (
    <PhoneShell title="Daftar & Pilih Peran" noNav>
      <div className="body pad auth-page">
        <p className="lead">Pilih peran bisnis, lalu masuk untuk mulai membuat profil, posting, dan mencari mitra.</p>
        <div className="role-grid">
          {[
            ['BUYER', '🛒', 'Pembeli', 'Posting kebutuhan bahan pangan.'],
            ['PRODUCER', '🌾', 'Produsen', 'Tawarkan supply dan kapasitas produksi.'],
          ].map(([value, icon, title, desc]) => (
            <button key={value} className={`rolecard ${role === value ? 'selected' : ''}`} onClick={() => setRole(value)}>
              <span className="em">{icon}</span>
              <span><b>{title}</b><small>{desc}</small></span>
              <span className="arr">›</span>
            </button>
          ))}
        </div>
        <form onSubmit={submit} className="card form-card">
          <div className="segmented">
            <button type="button" className={mode === 'register' ? 'on' : ''} onClick={() => setMode('register')}>Daftar</button>
            <button type="button" className={mode === 'login' ? 'on' : ''} onClick={() => setMode('login')}>Masuk</button>
          </div>
          <Field label="Email">
            <input value={email} type="email" required onChange={(event) => setEmail(event.target.value)} />
          </Field>
          <Field label="Password">
            <input value={password} type="password" required minLength={8} onChange={(event) => setPassword(event.target.value)} />
          </Field>
          {error ? <div className="state-card error compact">{error}</div> : null}
          <Button loading={loading}>{mode === 'register' ? 'Buat Akun' : 'Masuk'}</Button>
        </form>
      </div>
    </PhoneShell>
  );
}
