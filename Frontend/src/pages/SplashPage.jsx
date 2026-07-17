import { Link } from 'react-router-dom';

export function SplashPage() {
  return (
    <main className="standalone-phone">
      <section className="splash">
        <span className="halo" />
        <span className="halo h2" />
        <div className="ali-big">🌱</div>
        <div className="wordmark">
          <h1>Jal<span>in</span></h1>
          <p>FoodLink AI · Temukan mitra pangan paling cocok</p>
        </div>
        <Link className="start-btn" to="/role">Mulai Sekarang</Link>
      </section>
    </main>
  );
}
