import { Link } from 'react-router-dom';
import logoJalin from '../assets/logo_Jalin.png';

export function SplashPage() {
  return (
    <main className="standalone-phone">
      <section className="splash">
        <img className="splash-logo" src={logoJalin} alt="Logo Jalin" />
        <div className="wordmark">
          <h1><span className="edge">J</span><span className="ali">ali</span><span className="edge">n</span></h1>
          <p>Match Better. Grow Together</p>
        </div>
        <Link className="start-btn" to="/role">Mulai Sekarang</Link>
      </section>
    </main>
  );
}
