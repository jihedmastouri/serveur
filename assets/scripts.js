function copyCode(id) {
  if (event.target.tagName === 'INPUT') {
    return;
  }
  const liElm = document.getElementById(id);
  const copyText = liElm.querySelector('code > span');
  navigator.clipboard.writeText(copyText.innerHTML);

  clearTimeout(window.copiedTimemout);
  document.getElementById('modal-text-copied').style.display = 'none';
  document.getElementById('modal-text-copied').style.display = 'block';

  window.copiedTimemout = setTimeout(() => {
    document.getElementById('modal-text-copied').style.display = 'none';
  }, 500);
}

window.addEventListener('click', (event) => {
  const side = document.querySelector('side');
  if (window.sideOpened && !side.contains(event.currenttarget) && !side.contains(event.target) && side.style.width !== '0') {
    closeSide();
  }
});

function prettify(details) {
  const res = document.getElementById('res-target');
  res.innerHTML = "<pre>" + JSON.stringify(JSON.parse(res.innerHTML), null, 4) + "</pre>";
}

function openSide() {
  const side = document.querySelector('side');
  const overlay = document.getElementById('overlay-side');
  overlay?.style.display = 'block';

  setTimeout(() => {
    window.sideOpened = true;
  }, 100);
  side.style.width = '50%';
  side.style.minWidth = '300px';
}

function closeSide() {
  const side = document.querySelector('side');
  const overlay = document.getElementById('overlay-side');
  overlay?.style.display = 'none';
  side.style.width = '0';
  side.style.minWidth = '0';
}

function onIdChange(id) {
  const el = document.getElementById(id);
  const baseUrl = el.querySelector('button').getAttribute('data-hx-get');
  const val = el.querySelector('input').value;
  const url = `${baseUrl}/${val}`;
  el.setAttribute('hx-get', url);
}
