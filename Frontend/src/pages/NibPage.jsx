import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { PhoneShell } from '../components/layout/PhoneShell.jsx';
import { Button } from '../components/common/Button.jsx';
import { Field } from '../components/forms/Field.jsx';
import { api } from '../services/api.js';

export function NibPage() {
  const navigate = useNavigate();
  const [nib, setNib] = useState('13092200451788');
  const [status, setStatus] = useState('idle');
  const [error, setError] = useState('');

  async function submit(event) {
    event.preventDefault();
    setStatus('loading');
    setError('');
    try {
      await api.submitVerification({ nib_number: nib });
      setStatus('success');
      setTimeout(() => navigate('/home'), 600);
    } catch (err) {
      setError(err.message);
      setStatus('error');
    }
  }

  return (
    <PhoneShell title="Verifikasi NIB" backTo="/profile-setup" noNav>
      <form className="body pad center-flow" onSubmit={submit}>
        <div className="nib-illu">🛡️</div>
        <h3>NIB opsional untuk membangun trust</h3>
        <p>Verifikasi membantu calon mitra melihat bisnis kamu lebih kredibel.</p>
        <Field label="Nomor NIB"><input value={nib} required onChange={(e) => setNib(e.target.value)} /></Field>
        {status === 'success' ? <div className="state-card success compact">NIB terkirim untuk diverifikasi.</div> : null}
        {error ? <div className="state-card error compact">{error}</div> : null}
        <Button loading={status === 'loading'}>Verifikasi NIB</Button>
        <button type="button" className="skip" onClick={() => navigate('/home')}>Lewati dulu</button>
      </form>
    </PhoneShell>
  );
}
