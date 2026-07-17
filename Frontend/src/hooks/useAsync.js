import { useCallback, useEffect, useState } from 'react';

export function useAsync(loader, deps = []) {
  const [data, setData] = useState(null);
  const [status, setStatus] = useState('idle');
  const [error, setError] = useState('');

  const run = useCallback(async () => {
    setStatus('loading');
    setError('');
    try {
      const result = await loader();
      setData(result);
      setStatus('success');
      return result;
    } catch (err) {
      setError(err.message || 'Gagal memuat data.');
      setStatus('error');
      return null;
    }
  }, deps);

  useEffect(() => {
    run();
  }, [run]);

  return { data, status, error, reload: run, setData };
}
