import { useEffect, useRef } from 'react';

const Turnstile = ({ siteKey, onVerify, onError, onExpire, size = 'flexible', theme = 'auto' }) => {
  const containerRef = useRef(null);
  const widgetIdRef = useRef(null);
  const scriptLoadedRef = useRef(false);

  useEffect(() => {
    if (!siteKey || !containerRef.current) return;

    const loadTurnstile = () => {
      if (window.turnstile && containerRef.current && widgetIdRef.current === null) {
        try {
          widgetIdRef.current = window.turnstile.render(containerRef.current, {
            sitekey: siteKey,
            callback: (token) => {
              if (onVerify) onVerify(token);
            },
            'error-callback': () => {
              if (onError) onError();
            },
            'expired-callback': () => {
              if (onExpire) onExpire();
            },
            theme: theme,
            size: size, // 'normal', 'compact', 'flexible'
          });
        } catch (error) {
          console.error('Turnstile render error:', error);
        }
      }
    };

    // Check if script is already loaded
    if (window.turnstile) {
      loadTurnstile();
      return;
    }

    // Avoid loading script multiple times
    if (scriptLoadedRef.current) return;

    // Check if script already exists in DOM
    const existingScript = document.querySelector('script[src*="turnstile"]');
    if (existingScript) {
      scriptLoadedRef.current = true;
      existingScript.addEventListener('load', loadTurnstile);
      return () => {
        existingScript.removeEventListener('load', loadTurnstile);
      };
    }

    // Load Turnstile script
    const script = document.createElement('script');
    script.src = 'https://challenges.cloudflare.com/turnstile/v0/api.js';
    script.async = true;
    script.defer = true;
    scriptLoadedRef.current = true;
    
    script.onload = loadTurnstile;
    script.onerror = () => {
      console.error('Failed to load Turnstile script');
      scriptLoadedRef.current = false;
    };

    document.head.appendChild(script);

    return () => {
      if (window.turnstile && widgetIdRef.current !== null) {
        try {
          window.turnstile.remove(widgetIdRef.current);
        } catch (error) {
          console.error('Turnstile remove error:', error);
        }
        widgetIdRef.current = null;
      }
    };
  }, [siteKey, onVerify, onError, onExpire]);

  if (!siteKey) return null;

  return (
    <div className="my-4 w-full">
      <div ref={containerRef} className="w-full"></div>
    </div>
  );
};

export default Turnstile;
