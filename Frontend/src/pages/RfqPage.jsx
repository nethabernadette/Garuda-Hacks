import { Link } from 'react-router-dom';
import { PhoneShell } from '../components/layout/PhoneShell.jsx';
import { Feedback } from '../components/common/Feedback.jsx';
import { api } from '../services/api.js';
import { useAsync } from '../hooks/useAsync.js';

const fallbackSummary = {
  document_number: 'RFQ-DEMO-2026-0714',
  buyer_company: 'UD Berkah Pangan',
  producer_company: 'CV Tani Makmur',
  product_list: [{
    product_name: 'Cabai Merah Keriting, Grade A',
    quantity: 600,
    unit: 'kg',
    unit_price: 37000,
    currency: 'IDR',
    total_value: 22200000,
    specifications: 'Grade A, segar, kemasan karung bersih',
    additional_notes: 'Pengiriman Senin pukul 06.00 WIB',
  }],
  total_value: 22200000,
  currency: 'IDR',
  payment_terms: 'Termin mingguan setelah barang diterima',
  delivery_address: 'Bandung, Jawa Barat',
  agreement_status: 'DEMO',
};

export function RfqPage() {
  const agreementId = localStorage.getItem('jalin_agreement_id');
  const doc = useAsync(
    () => agreementId
      ? api.agreementDocument(agreementId)
      : Promise.resolve({ summary: fallbackSummary, document_number: fallbackSummary.document_number }),
    [agreementId],
  );
  const summary = doc.data?.summary || fallbackSummary;
  const first = summary.product_list?.[0] || fallbackSummary.product_list[0];
  const unitPrice = Number(first.unit_price || 0);
  const totalValue = Number(summary.total_value || first.total_value || 0);

  return (
    <PhoneShell title="RFQ & Procurement" backTo="/agreement">
      <div className="body">
        {doc.status === 'loading' ? <Feedback status={doc.status} empty={false} /> : null}
        {doc.status === 'error' ? (
          <div className="state-card compact">Menampilkan receipt RFQ demo sampai agreement dikonfirmasi kedua pihak.</div>
        ) : null}

        <section className="receipt-doc">
          <div className="receipt-top">
            <small>FoodLink RFQ Receipt</small>
            <b>{summary.document_number}</b>
            <span>{summary.agreement_status === 'CONFIRMED' ? 'Confirmed procurement summary' : 'Demo procurement preview'}</span>
          </div>

          <div className="receipt-parties">
            <div><small>Pembeli</small><b>{summary.buyer_company}</b></div>
            <div><small>Produsen</small><b>{summary.producer_company}</b></div>
          </div>

          <div className="receipt-line strong">
            <span>Nama Barang</span>
            <b>{first.product_name}</b>
          </div>
          <div className="receipt-grid">
            <div><small>Kuantitas</small><b>{first.quantity} {first.unit}</b></div>
            <div><small>Harga</small><b>Rp{unitPrice.toLocaleString('id-ID')} / {first.unit}</b></div>
            <div><small>Total</small><b>Rp{totalValue.toLocaleString('id-ID')}</b></div>
            <div><small>Mata Uang</small><b>{summary.currency || first.currency || 'IDR'}</b></div>
          </div>

          <div className="receipt-line">
            <span>Pengiriman</span>
            <b>{summary.delivery_address}</b>
          </div>
          <div className="receipt-line">
            <span>Pembayaran</span>
            <b>{summary.payment_terms}</b>
          </div>

          <div className="receipt-terms">
            <small>Persyaratan</small>
            <p>{first.specifications || 'Grade A, segar, kemasan karung bersih'}</p>
            <p>{first.additional_notes || 'Pengiriman Senin pukul 06.00 WIB'}</p>
          </div>

          <div className="receipt-total">
            <span>Grand Total</span>
            <b>Rp{totalValue.toLocaleString('id-ID')}</b>
          </div>
        </section>

        <div className="pad">
          <a className="btn btn-ghost" href={doc.data?.html ? makeHTMLBlob(doc.data.html) : '#'} download="procurement-summary.html">Unduh HTML RFQ</a>
          <Link className="btn btn-navy" to="/contact">Buka Kontak Mitra</Link>
        </div>

        <div className="card reveal-lock">
          <span className="lk">LOCK</span>
          <span><b>Kontak dibuka setelah sepakat</b><small>Demi keamanan, Ali hanya membuka kontak jika kedua pihak sudah menandatangani.</small></span>
          <span className="ali-mini">AI</span>
        </div>
      </div>
    </PhoneShell>
  );
}

function makeHTMLBlob(html) {
  return URL.createObjectURL(new Blob([html], { type: 'text/html' }));
}
