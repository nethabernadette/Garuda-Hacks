import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { PhoneShell } from '../components/layout/PhoneShell.jsx';
import { Button } from '../components/common/Button.jsx';
import { Field } from '../components/forms/Field.jsx';
import { api } from '../services/api.js';
import { useAuth } from '../context/AuthContext.jsx';

export function ProfileSetupPage() {
  const navigate = useNavigate();
  const auth = useAuth();
  const [form, setForm] = useState({
    company_name: auth.profile?.company_name || 'UD Berkah Pangan',
    phone: auth.profile?.phone || '+62 812-0000-0000',
    city: auth.profile?.city || 'Bandung, Jawa Barat',
    business_type: auth.profile?.business_type || 'Distributor pangan',
    product_category: auth.profile?.product_category || 'Hortikultura',
    delivery_area: auth.profile?.delivery_area || 'Bandung, Garut, Jakarta',
    purchase_frequency: auth.profile?.purchase_frequency || 'Mingguan',
  });
  const [consent, setConsent] = useState(false);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  function update(name, value) {
    setForm((current) => ({ ...current, [name]: value }));
  }

  async function submit(event) {
    event.preventDefault();
    if (!consent) {
      setError('Persetujuan kontak wajib dicentang.');
      return;
    }
    setLoading(true);
    setError('');
    try {
      const data = await api.updateProfile(form);
      auth.setProfile(data);
      navigate('/nib');
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }

  return (
    <PhoneShell title="Profil Bisnis" backTo="/role" noNav>
      <form className="body pad" onSubmit={submit}>
        <div className="nib-illu">🌱</div>
        <Field label="Nama Usaha"><input required value={form.company_name} onChange={(e) => update('company_name', e.target.value)} /></Field>
        <div className="grid2">
          <Field label="Telepon"><input required value={form.phone} onChange={(e) => update('phone', e.target.value)} /></Field>
          <Field label="Kota"><input required value={form.city} onChange={(e) => update('city', e.target.value)} /></Field>
        </div>
        <Field label="Jenis Bisnis"><input value={form.business_type} onChange={(e) => update('business_type', e.target.value)} /></Field>
        <Field label="Kategori Produk"><input value={form.product_category} onChange={(e) => update('product_category', e.target.value)} /></Field>
        <Field label="Area Kirim / Alamat Bisnis"><textarea value={form.delivery_area} onChange={(e) => update('delivery_area', e.target.value)} /></Field>
        <label className="checkline">
          <input type="checkbox" checked={consent} onChange={(e) => setConsent(e.target.checked)} />
          Saya setuju kontak bisnis dibuka setelah agreement dikonfirmasi kedua pihak.
        </label>
        {error ? <div className="state-card error compact">{error}</div> : null}
        <Button loading={loading}>Simpan Profil</Button>
      </form>
    </PhoneShell>
  );
}
