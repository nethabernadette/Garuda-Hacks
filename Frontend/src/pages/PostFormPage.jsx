import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { PhoneShell } from '../components/layout/PhoneShell.jsx';
import { Button } from '../components/common/Button.jsx';
import { Field } from '../components/forms/Field.jsx';
import { api } from '../services/api.js';
import { useAuth } from '../context/AuthContext.jsx';

export function PostFormPage() {
  const navigate = useNavigate();
  const { profile } = useAuth();
  const role = profile?.role || 'BUYER';
  const isBuyer = role === 'BUYER';
  const [form, setForm] = useState({
    product_name: 'Cabai Merah Keriting',
    category: 'Hortikultura',
    subcategory: 'Cabai',
    description: 'Grade A untuk pasokan mingguan.',
    quantity: 600,
    unit: 'kg',
    price_min: 36000,
    price_max: 38000,
    budget_min: 35000,
    budget_max: 39000,
    location: 'Garut',
    delivery_area: 'Bandung, Jakarta',
    delivery_location: 'Bandung',
    needed_date: '2026-07-25',
    frequency: 'weekly',
    status: isBuyer ? 'open' : 'active',
  });
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [loading, setLoading] = useState(false);

  function update(name, value) {
    setForm((current) => ({ ...current, [name]: value }));
  }

  async function submit(event) {
    event.preventDefault();
    setLoading(true);
    setError('');
    setSuccess('');
    try {
      if (isBuyer) {
        await api.createDemand({
          product_name: form.product_name,
          category: form.category,
          subcategory: form.subcategory,
          description: form.description,
          quantity: Number(form.quantity),
          unit: form.unit,
          budget_min: Number(form.budget_min),
          budget_max: Number(form.budget_max),
          delivery_location: form.delivery_location,
          needed_date: form.needed_date,
          frequency: form.frequency,
          status: 'open',
        });
      } else {
        await api.createSupply({
          product_name: form.product_name,
          category: form.category,
          subcategory: form.subcategory,
          description: form.description,
          quantity: Number(form.quantity),
          unit: form.unit,
          minimum_order_quantity: 50,
          price_min: Number(form.price_min),
          price_max: Number(form.price_max),
          location: form.location,
          delivery_area: form.delivery_area,
          availability_status: 'available',
          status: 'active',
        });
      }
      setSuccess('Posting terbit. Ali mulai mencari mitra.');
      setTimeout(() => navigate('/match'), 700);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }

  return (
    <PhoneShell title={isBuyer ? 'Posting Need' : 'Posting Supply'} backTo="/home">
      <form className="body pad" onSubmit={submit}>
        <Field label="Produk"><input required value={form.product_name} onChange={(e) => update('product_name', e.target.value)} /></Field>
        <div className="grid2">
          <Field label="Kategori"><input required value={form.category} onChange={(e) => update('category', e.target.value)} /></Field>
          <Field label="Subkategori"><input value={form.subcategory} onChange={(e) => update('subcategory', e.target.value)} /></Field>
        </div>
        <Field label="Deskripsi"><textarea value={form.description} onChange={(e) => update('description', e.target.value)} /></Field>
        <div className="grid2">
          <Field label="Quantity"><input type="number" min="1" required value={form.quantity} onChange={(e) => update('quantity', e.target.value)} /></Field>
          <Field label="Unit"><input required value={form.unit} onChange={(e) => update('unit', e.target.value)} /></Field>
        </div>
        {isBuyer ? (
          <>
            <div className="grid2">
              <Field label="Budget Min"><input type="number" min="0" value={form.budget_min} onChange={(e) => update('budget_min', e.target.value)} /></Field>
              <Field label="Budget Max"><input type="number" min="0" value={form.budget_max} onChange={(e) => update('budget_max', e.target.value)} /></Field>
            </div>
            <Field label="Lokasi Kirim"><input required value={form.delivery_location} onChange={(e) => update('delivery_location', e.target.value)} /></Field>
            <Field label="Tanggal Dibutuhkan"><input type="date" value={form.needed_date} onChange={(e) => update('needed_date', e.target.value)} /></Field>
          </>
        ) : (
          <>
            <div className="grid2">
              <Field label="Harga Min"><input type="number" min="0" value={form.price_min} onChange={(e) => update('price_min', e.target.value)} /></Field>
              <Field label="Harga Max"><input type="number" min="0" value={form.price_max} onChange={(e) => update('price_max', e.target.value)} /></Field>
            </div>
            <Field label="Lokasi"><input required value={form.location} onChange={(e) => update('location', e.target.value)} /></Field>
            <Field label="Area Kirim"><input value={form.delivery_area} onChange={(e) => update('delivery_area', e.target.value)} /></Field>
          </>
        )}
        {error ? <div className="state-card error compact">{error}</div> : null}
        {success ? <div className="state-card success compact">{success}</div> : null}
        <Button loading={loading}>Terbitkan Posting</Button>
      </form>
    </PhoneShell>
  );
}
