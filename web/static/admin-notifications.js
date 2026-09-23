document.addEventListener('DOMContentLoaded', function () {
  var bell = document.getElementById('notif-bell');
  if (!bell) return;

  var MAX_ITEMS = 20;
  var TOAST_MS = 8000;
  var INITIAL_BACKOFF_MS = 1000;
  var MAX_BACKOFF_MS = 30000;

  var toggle = document.getElementById('notif-toggle');
  var badge = document.getElementById('notif-badge');
  var menu = document.getElementById('notif-menu');
  var list = document.getElementById('notif-list');
  var empty = document.getElementById('notif-empty');
  var toasts = document.getElementById('notif-toasts');
  var unread = 0;
  var backoff = INITIAL_BACKOFF_MS;

  loadRecent().then(connect);

  toggle.addEventListener('click', function () {
    menu.classList.toggle('show');
    unread = 0;
    renderBadge();
  });
  document.addEventListener('click', function (event) {
    if (!bell.contains(event.target)) menu.classList.remove('show');
  });

  function loadRecent() {
    return fetch(bell.dataset.recentUrl, { headers: { Accept: 'application/json' } }).then(function (response) {
      if (!response.ok) throw new Error(response.statusText);
      return response.json();
    }).then(function (items) {
      list.textContent = '';
      items.forEach(function (item) { list.appendChild(renderItem(item)); });
      renderEmpty();
    }).catch(function () {});
  }

  function connect() {
    var scheme = location.protocol === 'https:' ? 'wss://' : 'ws://';
    var socket = new WebSocket(scheme + location.host + bell.dataset.wsPath);
    socket.onopen = function () { backoff = INITIAL_BACKOFF_MS; };
    socket.onmessage = function (event) {
      var item;
      try { item = JSON.parse(event.data); } catch (err) { return; }
      list.insertBefore(renderItem(item), list.firstChild);
      while (list.children.length > MAX_ITEMS) list.removeChild(list.lastChild);
      renderEmpty();
      unread += 1;
      renderBadge();
      showToast(item);
    };
    socket.onclose = function () {
      setTimeout(connect, backoff);
      backoff = Math.min(backoff * 2, MAX_BACKOFF_MS);
    };
  }

  function renderItem(item) {
    var node = document.createElement('a');
    node.className = 'list-group-item list-group-item-action small';
    node.href = seatMapURL(item);

    var place = document.createElement('strong');
    place.className = 'd-block';
    place.textContent = placeText(item);

    var show = document.createElement('div');
    show.textContent = item.movieTitle + ' · ' + item.startsAt;

    var detail = document.createElement('div');
    detail.className = 'text-secondary';
    detail.textContent = seatsText(item) + ' · ' + item.total + ' · ' + item.userEmail;

    node.appendChild(place);
    node.appendChild(show);
    node.appendChild(detail);
    return node;
  }

  function renderEmpty() {
    empty.classList.toggle('d-none', list.children.length > 0);
  }

  function renderBadge() {
    badge.textContent = unread;
    badge.classList.toggle('d-none', unread === 0);
  }

  function showToast(item) {
    var toast = document.createElement('div');
    toast.className = 'toast show';
    toast.setAttribute('role', 'status');
    var body = document.createElement('a');
    body.className = 'toast-body d-block text-reset text-decoration-none';
    body.href = seatMapURL(item);
    var title = document.createElement('strong');
    title.className = 'd-block';
    title.textContent = 'New booking · ' + placeText(item);
    var detail = document.createElement('div');
    detail.textContent = item.movieTitle + ' · ' + seatsText(item);
    body.appendChild(title);
    body.appendChild(detail);
    toast.appendChild(body);
    toasts.appendChild(toast);
    setTimeout(function () { toast.remove(); }, TOAST_MS);
  }

  function seatMapURL(item) {
    return bell.dataset.seatMapUrl.replace(':id', encodeURIComponent(String(item.showtimeId)));
  }

  function placeText(item) {
    return item.theater + ' · ' + item.room;
  }

  function seatsText(item) {
    return (item.seats || []).join(', ');
  }
});
