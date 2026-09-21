document.addEventListener('DOMContentLoaded', function () {
  document.querySelectorAll('.seat[data-dialog]').forEach(function (seat) {
    seat.addEventListener('click', function () {
      var dialog = document.getElementById(seat.dataset.dialog);
      if (!dialog) return;
      dialog.showModal();
      var input = dialog.querySelector('input[name=code]');
      if (input) input.focus();
    });
  });

  var form = document.getElementById('sell-form');
  var summary = document.getElementById('sell-summary');
  var submit = document.getElementById('sell-submit');
  if (!form || !summary || !submit) return;
  var checks = form.querySelectorAll('.seat-check');

  function refresh() {
    var count = 0;
    var total = 0;
    checks.forEach(function (check) {
      if (!check.checked) return;
      count++;
      total += Number(check.dataset.price || 0);
    });
    summary.textContent = count + ' seat' + (count === 1 ? '' : 's') + ' · ' + total.toFixed(2);
    submit.disabled = count === 0;
  }

  checks.forEach(function (check) { check.addEventListener('change', refresh); });
  refresh();
});
