// TODO: Replace with backend API when match recommendation discovery endpoints are available.
export const demoMatches = [
  {
    id: 'match-demo-1',
    company: 'CV Tani Makmur',
    city: 'Garut, Jawa Barat',
    product: 'Cabai Merah Keriting',
    score: 92,
    category: 'Hortikultura',
    price: 'Rp37.000/kg',
    capacity: '600 kg/minggu',
    reason: 'Kategori, volume, dan area kirim cocok dengan kebutuhan pembeli.',
  },
  {
    id: 'match-demo-2',
    company: 'Koperasi Sari Bumi',
    city: 'Lembang',
    product: 'Tomat Grade A',
    score: 86,
    category: 'Sayuran',
    price: 'Rp12.500/kg',
    capacity: '1 ton/minggu',
    reason: 'Kapasitas besar dengan jadwal panen mingguan.',
  },
];

// TODO: Replace with backend chat list API when room discovery endpoint is available without a match ID.
export const demoMessages = [
  { id: 1, side: 'in', text: 'Halo Bu Sari, kami bisa supply cabai 600 kg/minggu.', time: '18:40' },
  { id: 2, side: 'out', text: 'Bisa kirim setiap Senin pagi ke Bandung?', time: '18:42' },
  { id: 3, side: 'in', text: 'Bisa. Harga Rp37.000/kg untuk Grade A.', time: '18:45' },
];

export const demoAgreement = {
  match_id: 'match-demo-1',
  items: [
    {
      product_name: 'Cabai Merah Keriting, Grade A',
      quantity: 600,
      unit: 'kg',
      unit_price: 37000,
      currency: 'IDR',
      delivery_date: '2026-07-25',
      delivery_address: 'Gudang UD Berkah Pangan, Bandung',
      payment_terms: 'Termin mingguan setelah barang diterima',
      specification: 'Grade A, segar, kemasan karung bersih',
      additional_notes: 'Pengiriman Senin pukul 06.00 WIB',
    },
  ],
};
