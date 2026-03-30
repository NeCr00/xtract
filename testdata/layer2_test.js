// ============================================
// Layer 2: AST-Based Extraction Test Data
// ============================================

// Technique 15: fetch() calls
fetch('/api/v1/users');
fetch('/api/v1/posts', { method: 'POST', body: JSON.stringify({ title, content }) });
fetch("https://api.example.com/data", {
    method: "PUT",
    headers: { "Content-Type": "application/json", "Authorization": "Bearer " + token },
    body: JSON.stringify({ name: "test" })
});
fetch(`/api/v2/items/${itemId}`);

// Technique 16: XMLHttpRequest
var xhr = new XMLHttpRequest();
xhr.open('GET', '/api/v1/status');
xhr.open("POST", "/api/v1/submit");
xhr.open('DELETE', 'https://api.example.com/v1/resource/123');

// Technique 17: axios calls
axios.get('/api/users');
axios.post('/api/users', { name: 'John' });
axios.put('/api/users/123', data);
axios.delete('/api/users/123');
axios.patch('/api/users/123', { name: 'Jane' });
axios.request({ url: '/api/config', method: 'get' });
axios({ url: '/api/health', method: 'HEAD' });

// Technique 18: jQuery AJAX
$.ajax({ url: '/api/legacy/data', type: 'GET' });
$.get('/api/legacy/users');
$.post('/api/legacy/submit', data);
$.getJSON('/api/legacy/config.json');
jQuery.ajax({ url: '/api/jquery/endpoint', method: 'POST' });

// Technique 19: navigator.sendBeacon
navigator.sendBeacon('/api/analytics/track', data);
navigator.sendBeacon("https://analytics.example.com/collect");

// Technique 20: EventSource / WebSocket
var es = new EventSource('/api/events/stream');
var ws = new WebSocket('wss://realtime.example.com/ws');
var ws2 = new WebSocket("ws://dev.example.com:8080/socket");

// Technique 21: Dynamic import()
import('./modules/dashboard.js');
import("./components/LazyWidget");
import(`./locale/${lang}.js`);

// Technique 22: require() calls
require('./utils/helpers');
require('../lib/database');
require('./config/settings.json');

// Technique 23: document.location assignments
document.location = '/login';
document.location.href = '/dashboard';
window.location = 'https://example.com/redirect';
window.location.href = '/new-page';
location.href = '/fallback';
location.assign('/assigned-page');
location.replace('/replaced-page');

// Technique 24: window.open()
window.open('https://docs.example.com/help');
window.open('/popup/details', '_blank');
open('/legacy/popup');

// Technique 25: Element .src / .href assignments
img.src = '/images/avatar.png';
link.href = '/styles/main.css';
script.src = 'https://cdn.example.com/lib.js';
video.src = '/media/intro.mp4';

// Technique 26: setAttribute
el.setAttribute('src', '/images/banner.jpg');
el.setAttribute("href", "/pages/about");
form.setAttribute('action', '/api/submit');

// Technique 27: innerHTML with URLs
div.innerHTML = '<a href="/inner/link">Click</a><img src="/inner/image.png">';
container.outerHTML = '<script src="/inner/script.js"></script>';

// Technique 28: postMessage
targetWindow.postMessage(data, 'https://trusted.example.com');
iframe.contentWindow.postMessage({type: 'init'}, 'https://widget.example.com');

// Technique 29: Form action in JS HTML
var formHtml = '<form action="/api/form/submit" method="POST">';
document.write('<form action="/legacy/form/handler">');

// Technique 30: Service Worker
navigator.serviceWorker.register('/sw.js');
navigator.serviceWorker.register("./service-worker.js");

// Technique 31: Web Worker
var worker = new Worker('/workers/compute.js');
var shared = new SharedWorker('/workers/shared-state.js');

// Technique 32: Webpack require
require.ensure([], function(require) { require('./chunk-a'); });
require.context('./templates', true, /\.html$/);
__webpack_require__(42);

// Technique 33: Dynamic script loading
var script = document.createElement('script');
script.src = 'https://cdn.example.com/analytics.js';
document.head.appendChild(script);

var s = document.createElement("script");
s.src = "/js/tracking.js";
document.body.appendChild(s);
