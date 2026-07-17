import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { PhoneShell } from '../components/layout/PhoneShell.jsx';
import { Button } from '../components/common/Button.jsx';
import { api } from '../services/api.js';
import { demoAgreement } from '../data/demoData.js';

export function AgreementPage() {
  const navigate = useNavigate();
  const [agreementId, setAgreementId] = useState(localStorage.getItem('jalin_agreement_id') || '');
  const [status, setStatus] = useState('idle');
  const [error, setError] = useState('');

  async function createAgreement() {
    setStatus('loading');
    setError('');
    try {
      const data = await api.createAgreement(demoAgreement);
      const id = data?.id || data?.agreement_id;
      if (id) {
        localStorage.setItem('jalin_agreement_id', id);
        setAgreementId(id);
      }
      setStatus('success');
    } catch (err) {
      setError(err.message);
      setStatus('error');
    }
  }

  async function confirmAgreement() {
    if (!agreementId) {
      await createAgreement();
      return;
    }
    setStatus('loading');
    try {
      await api.confirmAgreement(agreementId);
      setStatus('success');
      navigate('/rfq');
    } catch (err) {
      setError(err.message);
      setStatus('error');
    }
  }

  return (
    <PhoneShell title="Form Kesepakatan" backTo="/chat">
      <div className="body pad">
        <div className="steps">
          <span className="stp done"><i>✓</i><b>Match</b></span>
          <span className="stp done"><i>✓</i><b>Nego</b></span>
          <span className="stp"><i>3</i><b>Sepakat</b></span>
        </div>
        <article className="card agreement-card">
          <h3>Kesepakatan Supply Cabai</h3>
          <div className="frow"><small>Produk</small><b>Cabai Merah Keriting, Grade A</b></div>
          <div className="frow"><small>Volume</small><b>600 kg / minggu</b></div>
          <div className="frow"><small>Harga</small><b>Rp37.000 / kg</b></div>
          <div className="frow"><small>Pengiriman</small><b>Senin 06.00 WIB ke Bandung</b></div>
          <div className="frow"><small>Pembayaran</small><b>Termin mingguan setelah barang diterima</b></div>
        </article>
        {agreementId ? <div className="state-card success compact">Draft agreement tersimpan: {agreementId}</div> : null}
        {error ? <div className="state-card error compact">{error}</div> : null}
        <Button loading={status === 'loading'} onClick={agreementId ? confirmAgreement : createAgreement}>
          {agreementId ? 'Konfirmasi Kesepakatan' : 'Buat Agreement'}
        </Button>
        <Button variant="ghost" onClick={() => navigate('/rfq')}>Lihat RFQ Demo</Button>
      </div>
    </PhoneShell>
  );
}
