import { PhoneShell } from '../components/layout/PhoneShell.jsx';
import { Feedback } from '../components/common/Feedback.jsx';
import { Button } from '../components/common/Button.jsx';
import { api } from '../services/api.js';
import { useAsync } from '../hooks/useAsync.js';

export function NotificationsPage() {
  const list = useAsync(() => api.notifications({ limit: 20 }), []);
  const items = list.data?.items || [];

  async function markAll() {
    await api.markAllNotificationsRead();
    list.reload();
  }

  return (
    <PhoneShell title="Notifikasi" backTo="/home">
      <div className="body pad">
        <Button variant="ghost" onClick={markAll}>Tandai semua dibaca</Button>
        <Feedback status={list.status} error={list.error} empty={list.status === 'success' && items.length === 0}>
          {items.map((item) => (
            <article key={item.id} className={`card nrow ${item.is_read ? '' : 'unread'}`}>
              <span className="em">🔔</span>
              <span>
                <b>{item.title}</b>
                <p>{item.message}</p>
              </span>
            </article>
          ))}
        </Feedback>
      </div>
    </PhoneShell>
  );
}
