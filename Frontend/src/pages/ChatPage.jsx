import { useState } from 'react';
import { Link } from 'react-router-dom';
import { PhoneShell } from '../components/layout/PhoneShell.jsx';
import { demoMessages } from '../data/demoData.js';

export function ChatPage() {
  const [messages, setMessages] = useState(demoMessages);
  const [text, setText] = useState('');

  function send(event) {
    event.preventDefault();
    if (!text.trim()) return;
    setMessages((current) => [...current, { id: Date.now(), side: 'out', text, time: '19:02' }]);
    setText('');
  }

  return (
    <PhoneShell title="" noNav>
      <header className="chat-head">
        <Link className="back-btn" to="/match">‹</Link>
        <span className="pf">🌶️</span>
        <span className="nm"><b>CV Tani Makmur</b><small>Online · NIB terverifikasi</small></span>
      </header>
      <div className="pin">
        <span>📌</span>
        <span><b>Draft Kesepakatan</b><small>Harga dan jadwal hampir lengkap</small></span>
        <Link className="open" to="/agreement">Buka</Link>
      </div>
      <div className="msgs">
        {messages.map((msg) => (
          <div key={msg.id} className={`msg ${msg.side}`}>
            {msg.text}<time>{msg.time}</time>
          </div>
        ))}
        <div className="offer">
          <small>PENAWARAN</small>
          <div className="rw"><span>Cabai Grade A</span><b>Rp37.000/kg</b></div>
          <div className="rw"><span>Volume</span><b>600 kg/minggu</b></div>
          <div className="ok">
            <Link className="y" to="/agreement">Setujui</Link>
            <button className="n" onClick={() => setMessages((m) => [...m, { id: Date.now(), side: 'out', text: 'Bisa Rp36.800/kg?', time: '19:03' }])}>Nego</button>
          </div>
        </div>
      </div>
      <form className="inputbar" onSubmit={send}>
        <input value={text} onChange={(e) => setText(e.target.value)} placeholder="Tulis pesan..." />
        <button className="send">➤</button>
      </form>
    </PhoneShell>
  );
}
