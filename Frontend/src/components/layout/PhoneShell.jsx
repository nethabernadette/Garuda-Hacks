import { NavLink, useLocation } from 'react-router-dom';

const navItems = [
  { to: '/home', label: 'Beranda', icon: '⌂' },
  { to: '/match', label: 'Match', icon: '✦' },
  { to: '/post', label: '', icon: '+' },
  { to: '/chat', label: 'Chat', icon: '◌' },
  { to: '/profile', label: 'Profil', icon: '◍' },
];

export function PhoneShell({ children, title = 'Jalin', backTo, noNav = false, gradient = false }) {
  const location = useLocation();
  const activePost = location.pathname === '/post';

  return (
    <main className="app-stage">
      <div className="brandbar">
        <span className="brand-mark">J</span>
        <div>
          <b>Jalin <i>FoodLink AI</i></b>
          <span> B2B pangan, dari match sampai sepakat</span>
        </div>
      </div>
      <div className="phone">
        <div className="notch" />
        <section className="screen on">
          <StatusBar light={gradient} />
          {title && (
            <header className={gradient ? 'gradhead' : 'plainhead'}>
              {backTo ? <NavLink className={gradient ? 'round-btn' : 'back-btn'} to={backTo}>‹</NavLink> : <span className="head-spacer" />}
              <h2>{title}</h2>
              <span className="head-spacer" />
            </header>
          )}
          {children}
          {!noNav && (
            <nav className="nav">
              {navItems.map((item) => (
                <NavLink key={item.to} to={item.to} className={({ isActive }) => `${isActive || (item.to === '/post' && activePost) ? 'on' : ''} ${item.to === '/post' ? 'plus-link' : ''}`}>
                  {item.to === '/post' ? <span className="plus">{item.icon}</span> : <span className="nav-ic">{item.icon}</span>}
                  {item.label}
                </NavLink>
              ))}
            </nav>
          )}
        </section>
      </div>
      <aside className="panel">
        <h4>✦ Peta Layar</h4>
        <p>Prototype React mempertahankan bentuk aplikasi HP dari HTML awal. Buka di desktop tetap dibatasi seperti layar mobile.</p>
        <div className="chips">
          {[
            ['/home', 'Beranda'],
            ['/post', 'Posting'],
            ['/match', 'AI Match'],
            ['/notifications', 'Notifikasi'],
            ['/chat', 'Chat'],
            ['/agreement', 'Kesepakatan'],
            ['/rfq', 'RFQ'],
            ['/contact', 'Kontak'],
            ['/profile', 'Profil'],
          ].map(([to, label]) => (
            <NavLink key={to} to={to}>{label}</NavLink>
          ))}
        </div>
        <div className="legend">
          <b>Alur:</b> profil bisnis → post need/supply → match → chat → agreement → RFQ → kontak terbuka.
        </div>
      </aside>
    </main>
  );
}

function StatusBar({ light }) {
  return (
    <div className={`sb ${light ? 'light' : ''}`}>
      <span>19:02</span>
      <span className="ic">▰ LTE ▱</span>
    </div>
  );
}
