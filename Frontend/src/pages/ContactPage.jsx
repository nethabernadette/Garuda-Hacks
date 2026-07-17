import { PhoneShell } from '../components/layout/PhoneShell.jsx';
import { Feedback } from '../components/common/Feedback.jsx';
import { api } from '../services/api.js';
import { useAsync } from '../hooks/useAsync.js';

const fallbackContact = {
  producer: {
    company_name: 'CV Tani Makmur',
    phone_number: '+62 812-2043-8871',
    email: 'halo@tanimakmur.co.id',
    business_address: 'Jl. Raya Cikajang KM 4, Garut, Jawa Barat',
    business_representative: 'Pak Dedi',
  },
};

export function ContactPage() {
  const agreementId = localStorage.getItem('jalin_agreement_id');
  const contact = useAsync(
    () => agreementId ? api.agreementContact(agreementId) : Promise.resolve(fallbackContact),
    [agreementId],
  );
  const producer = contact.data?.producer || fallbackContact.producer;

  return (
    <PhoneShell title="Kontak Terbuka 🔓" gradient backTo="/rfq">
      <div className="body">
        <Feedback status={contact.status} error={contact.error} empty={false}>
          <div className="card contact-card">
            <span className="pf">🌶️</span>
            <h3>{producer.company_name}</h3>
            <span className="badge-nib">✓ NIB Terverifikasi</span>
            <Info label="Telepon / WhatsApp" value={`${producer.phone_number || '-'} ${producer.business_representative ? `(${producer.business_representative})` : ''}`} icon="☎" />
            <Info label="Email" value={producer.email || '-'} icon="✉" />
            <Info label="Alamat Gudang" value={producer.business_address || '-'} icon="⌖" />
            <a className="btn btn-teal" href={`https://wa.me/${(producer.phone_number || '').replace(/\D/g, '')}`}>💬 Hubungi via WhatsApp</a>
          </div>
        </Feedback>
        {contact.status === 'error' ? <div className="state-card compact">Kontak asli belum terbuka dari backend. Tampilan contoh tetap dipertahankan.</div> : null}
        <div className="card ali-note">
          <span className="ali-mini">🌱</span>
          <p><b>Ali:</b> Kerja sama pertama kalian dimulai! Pembayaran dan pengiriman kalian atur langsung ya.</p>
        </div>
      </div>
    </PhoneShell>
  );
}

function Info({ label, value, icon }) {
  return (
    <div className="crow">
      <span>{icon}</span>
      <span><small>{label}</small><b>{value}</b></span>
    </div>
  );
}
