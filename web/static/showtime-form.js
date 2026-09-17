document.addEventListener('DOMContentLoaded', function () {
  var movie = document.getElementById('movie_id');
  var starts = document.getElementById('starts_at');
  var ends = document.getElementById('ends_at');
  var hint = document.getElementById('movie-ends-hint');
  if (!movie || !starts || !ends || !hint) return;

  function pad(n) { return (n < 10 ? '0' : '') + n; }

  function local(d) {
    return d.getFullYear() + '-' + pad(d.getMonth() + 1) + '-' + pad(d.getDate()) + 'T' + pad(d.getHours()) + ':' + pad(d.getMinutes());
  }

  function update(fill) {
    var option = movie.options[movie.selectedIndex];
    var duration = option ? parseInt(option.dataset.duration, 10) : NaN;
    var start = new Date(starts.value);
    if (!duration || isNaN(start.getTime())) { hint.textContent = ''; return; }
    var end = new Date(start.getTime() + duration * 60000);
    hint.textContent = 'Movie ends at ' + pad(end.getHours()) + ':' + pad(end.getMinutes()) + (end.getDate() !== start.getDate() ? ' (+1 day)' : '');
    if (!ends.readOnly && (fill || ends.value === '')) ends.value = local(end);
  }

  function refill() { update(true); }

  movie.addEventListener('change', refill);
  starts.addEventListener('change', refill);
  starts.addEventListener('input', refill);
  update(false);
});
