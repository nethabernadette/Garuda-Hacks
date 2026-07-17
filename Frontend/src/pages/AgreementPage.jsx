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

  async function ensureMatchId() {
    const stored = localStorage.getItem('jalin_match_id');
    if (stored) return stored;

    const existing = await api.matches({ limit: 1 });
    if (Array.isArray(existing) && existing[0]?.id) {
      localStorage.setItem('jalin_match_id', existing[0].id);
      return existing[0].id;
    }

    const feed = await api.feed({ type: 'supply', q: 'Cabai', limit: 10 });
    const items = Array.isArray(feed?.items) ? feed.items : [];
    const supply = items.find((item) => String(item.product_name || '').toLowerCase().includes('cabai')) || items[0];
    if (!supply?.id) {
      throw new Error('Belum ada supply post untuk dibuat match.');
    }

    const match = await api.createInterestMatch({ supply_post_id: supply.id });
    if (!match?.id) {
      throw new Error('Match belum berhasil dibuat.');
    }
    localStorage.setItem('jalin_match_id', match.id);
    return match.id;
  }

  async function createAgreement() {
    setStatus('loading');
    setError('');
    try {
      const matchId = await ensureMatchId();
      const data = await api.createAgreement({ ...demoAgreement, match_id: matchId });
      const id = data?.id || data?.agreement_id;
      if (id) {
        localStorage.setItem('jalin_agreement_id', id);
        setAgreementId(id);
      }
      setStatus('success');
    } catch (err) {
      if (/active agreement already exists/i.test(err.message)) {
        const agreements = await api.agreements();
        const active = Array.isArray(agreements)
          ? agreements.find((item) => item.match_id === matchId && item.status !== 'CANCELLED')
          : null;
        if (active?.id) {
          localStorage.setItem('jalin_agreement_id', active.id);
          setAgreementId(active.id);
          setStatus('success');
          setError('');
          return;
        }
      }
      setError(err.message);
      setStatus('error');
    }
  }

  async function resetDraftAndCreateAgreement() {
    localStorage.removeItem('jalin_agreement_id');
    localStorage.removeItem('jalin_match_id');
    setAgreementId('');
    await createAgreement();
  }

  async function confirmAgreement() {
    if (!agreementId) {
      await createAgreement();
      return;
    }
    setStatus('loading');
    try {
      const response = await api.confirmAgreement(agreementId);
      if (!response?.buyer_confirmed || !response?.producer_confirmed || response?.status !== 'CONFIRMED') {
        await api.demoConfirmAgreement(agreementId);
      }
      setStatus('success');
      navigate('/rfq');
    } catch (err) {
      if (err.status === 403 || /not part of this match/i.test(err.message)) {
        await resetDraftAndCreateAgreement();
        return;
      }
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
