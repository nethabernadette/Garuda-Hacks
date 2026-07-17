import { Link } from 'react-router-dom';
import { PhoneShell } from '../components/layout/PhoneShell.jsx';
import { Feedback } from '../components/common/Feedback.jsx';
import { api } from '../services/api.js';
import { useAsync } from '../hooks/useAsync.js';

const fallbackSummary = {
  document_number: 'RFQ-2026-0714',
  buyer_company: 'UD Berkah Pangan',
  producer_company: 'CV Tani Makmur',
  product_list: [{ product_name: 'Cabai Merah Keriting, Grade A', quantity: 600, unit: 'kg', unit_price: 37000, currency: 'IDR', total_value: 22200000 }],
  total_value: 22200000,
  currency: 'IDR',
  payment_terms: 'Termin mingguan setelah barang diterima',
  delivery_address: 'Bandung, Jawa Barat',
  agreement_status: 'CONFIRMED',
};

export function RfqPage() {
  const agreementId = localStorage.getItem('jalin_agreement_id');
  const doc = useAsync(
    () => agreementId ? api.agreementDocument(agreementId) : Promise.resolve({ summary: fallbackSummary, document_number: fallbackSummary.document_number }),
    [agreementId],
  );
  const summary = doc.data?.summary || fallbackSummary;
  const first = summary.product_list?.[0] || {};

  return (
    <PhoneShell title="RFQ & Procurement" backTo="/agreement">
      <div className="body">
        <Feedback status={doc.status} error={doc.error} empty={false}>
          <div className="rfq-doc">
            <div className="hd"><b>{summary.document_number}</b><span>Dibuat otomatis oleh Ali ✨</span></div>
            <div className="frow"><small>Pembeli</small><b>{summary.buyer_company}</b></div>
            <div className="frow"><small>Produsen</small><b>{summary.producer_company}</b></div>
            <div className="frow"><small>Item</small><b>{first.product_name}</b></div>
            <div className="frow"><small>Volume Total</small><b>{first.quantity} {first.unit}</b></div>
            <div className="frow"><small>Harga Satuan</small><b>Rp{Number(first.unit_price || 0).toLocaleString('id-ID')} / {first.unit}</b></div>
            <div className="frow"><small>Estimasi Nilai</small><b>Rp{Number(summary.total_value || 0).toLocaleString('id-ID')}</b></div>
            <div className="frow"><small>Status</small><b className="ok-text">✓ Disepakati kedua pihak</b></div>
          </div>
        </Feedback>
        {doc.status === 'error' ? <div className="state-card compact">Menampilkan contoh visual RFQ karena backend belum mengembalikan dokumen confirmed.</div> : null}
        <div className="pad">
          <a className="btn btn-ghost" href={doc.data?.html ? makeHTMLBlob(doc.data.html) : '#'} download="procurement-summary.html">⬇ Unduh HTML RFQ</a>
          <Link className="btn btn-navy" to="/contact">🔓 Buka Kontak Mitra</Link>
        </div>
        <div className="card reveal-lock">
          <span className="lk">🔒</span>
          <span><b>Kontak dibuka setelah sepakat</b><small>Demi keamanan, Ali hanya membuka kontak jika kedua pihak sudah menandatangani.</small></span>
          <span className="ali-mini">🌱</span>
        </div>
      </div>
    </PhoneShell>
  );
}

function makeHTMLBlob(html) {
  return URL.createObjectURL(new Blob([html], { type: 'text/html' }));
}
