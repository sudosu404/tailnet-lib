
self.addEventListener("install", (event) => {
  event.waitUntil(
    caches.open("app-cache").then((cache) => {
      return cache.addAll(["/index.css", "/index.js", "/icons/tsdproxy.svg", "/icons/sh/github.svg", "/icons/sh/x.svg", "/icons/icon-192x192.png", "icons/icon-512x512.png"]);
    })
  );
});

self.addEventListener("fetch", (event) => {
  event.respondWith(
    caches.match(event.request).then((response) => {
      return response || fetch(event.request);
    })
  );
});
