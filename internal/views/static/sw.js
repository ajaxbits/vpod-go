const CACHE_NAME = 'vpod-v1';
const STATIC_ASSETS = [
  '/ui/',
  '/ui/static/css/pico.min.css',
  '/ui/static/js/htmx@2.0.4.min.js',
  '/ui/static/manifest.json',
  '/ui/static/icons/icon-192.svg',
  '/ui/static/icons/icon-512.svg',
  '/ui/static/icons/icon-maskable.svg'
];

// Install event: cache static assets
self.addEventListener('install', (event) => {
  event.waitUntil(
    caches.open(CACHE_NAME).then((cache) => {
      return cache.addAll(STATIC_ASSETS);
    })
  );
  self.skipWaiting();
});

// Activate event: clean up old caches
self.addEventListener('activate', (event) => {
  event.waitUntil(
    caches.keys().then((cacheNames) => {
      return Promise.all(
        cacheNames
          .filter((name) => name !== CACHE_NAME)
          .map((name) => caches.delete(name))
      );
    })
  );
  self.clients.claim();
});

// Fetch event: network-first for API calls, cache-first for static assets
self.addEventListener('fetch', (event) => {
  const url = new URL(event.request.url);

  // Skip non-GET requests
  if (event.request.method !== 'GET') {
    return;
  }

  // For API/dynamic content (feeds, gen, etc.), use network-first
  if (url.pathname.startsWith('/ui/feeds') || 
      url.pathname.startsWith('/ui/gen') ||
      url.pathname.startsWith('/feed/')) {
    event.respondWith(
      fetch(event.request)
        .catch(() => caches.match(event.request))
    );
    return;
  }

  // For static assets, use cache-first
  if (url.pathname.startsWith('/ui/static/') || url.pathname === '/ui/') {
    event.respondWith(
      caches.match(event.request).then((cachedResponse) => {
        if (cachedResponse) {
          // Return cached version, but also update cache in background
          fetch(event.request).then((response) => {
            if (response.ok) {
              caches.open(CACHE_NAME).then((cache) => {
                cache.put(event.request, response);
              });
            }
          });
          return cachedResponse;
        }
        return fetch(event.request).then((response) => {
          if (response.ok) {
            const responseClone = response.clone();
            caches.open(CACHE_NAME).then((cache) => {
              cache.put(event.request, responseClone);
            });
          }
          return response;
        });
      })
    );
    return;
  }

  // Default: network only
  event.respondWith(fetch(event.request));
});
