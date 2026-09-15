document.addEventListener('DOMContentLoaded', function () {
  var map = document.querySelector('.seat-map');
  if (!map) return;

  map.addEventListener('click', function (event) {
    var seat = event.target.closest('.seat[data-url]');
    if (!seat) return;
    seat.disabled = true;
    fetch(seat.dataset.url, {
      method: 'POST',
      headers: { 'X-CSRF-Token': map.dataset.csrf },
    }).then(function (response) {
      if (!response.ok) throw new Error(response.statusText);
      return response.json();
    }).then(function (data) {
      seat.classList.toggle('seat-off', !data.is_active);
      seat.title = seat.title.replace(' · disabled', '') + (data.is_active ? '' : ' · disabled');
    }).catch(function () {
      showError(map, 'Could not update seat');
    }).finally(function () {
      seat.disabled = false;
    });
  });

  function showError(map, message) {
    var existing = document.getElementById('seat-alert');
    if (existing) existing.remove();
    var alert = document.createElement('div');
    alert.id = 'seat-alert';
    alert.className = 'alert alert-danger alert-dismissible py-2';
    alert.textContent = message;
    var close = document.createElement('button');
    close.type = 'button';
    close.className = 'btn-close';
    close.setAttribute('aria-label', 'Close');
    close.addEventListener('click', function () { alert.remove(); });
    alert.appendChild(close);
    map.insertAdjacentElement('beforebegin', alert);
  }
});
