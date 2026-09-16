document.addEventListener('DOMContentLoaded', function () {
  var rows = document.getElementById('cast-rows');
  var add = document.getElementById('add-cast');
  if (!rows || !add) return;

  add.addEventListener('click', function () {
    var last = rows.querySelector('.cast-row:last-child');
    var row = last.cloneNode(true);
    row.querySelectorAll('input').forEach(function (input) { input.value = ''; });
    rows.appendChild(row);
    row.querySelector('input').focus();
  });

  rows.addEventListener('click', function (event) {
    if (!event.target.classList.contains('remove-cast')) return;
    var row = event.target.closest('.cast-row');
    if (rows.querySelectorAll('.cast-row').length > 1) {
      row.remove();
    } else {
      row.querySelectorAll('input').forEach(function (input) { input.value = ''; });
    }
  });
});
