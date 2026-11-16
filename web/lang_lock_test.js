// A platform administrator can lock the interface language. boot.js must apply
// the lock before paint (ignoring the per-browser choice), i18n.js must refuse
// to switch, and both pages must hide their language switch.
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const assert = require('node:assert/strict');

const read = name => fs.readFileSync(path.join(__dirname, name), 'utf8');

function boot(policy, stored) {
  const writes = [];
  let reloads = 0;
  const root = {dataset: {}, lang: ''};
  const context = {
    window: {LF_LANG_POLICY: policy, confirm() {}, alert() {}, prompt() {}},
    localStorage: {getItem: () => stored, setItem: (k, v) => writes.push([k, v])},
    location: {reload: () => { reloads++; }},
    document: {documentElement: root, body: null, addEventListener() {}},
    MutationObserver: class { observe() {} },
  };
  context.window.localStorage = context.localStorage;
  context.window.location = context.location;
  vm.createContext(context);
  vm.runInContext('var window = this.window;', context);
  vm.runInContext(read('boot.js'), context);
  vm.runInContext('window.LF_I18N_EN = {};', context);
  vm.runInContext(read('i18n.js'), context);
  return {root, win: context.window, writes, reloads: () => reloads};
}

// Locked to English: the browser's stored Chinese choice is ignored.
let page = boot('en', 'zh');
assert.equal(page.root.dataset.lang, 'en');
assert.equal(page.root.lang, 'en');
assert.equal(page.root.dataset.langLocked, 'true');
assert.equal(page.win.lfLang(), 'en');
assert.equal(page.win.lfLangLocked(), true);
page.win.lfSetLang('zh');
assert.deepEqual(page.writes, []);
assert.equal(page.reloads(), 0);

// Locked to Chinese overrides a stored English choice.
page = boot('zh', 'en');
assert.equal(page.root.dataset.lang, 'zh');
assert.equal(page.root.dataset.langLocked, 'true');

// No lock (or an unknown value): each browser keeps its own choice and can switch.
for (const policy of [undefined, '', 'fr']) {
  page = boot(policy, 'en');
  assert.equal(page.root.dataset.lang, 'en');
  assert.equal(page.root.dataset.langLocked, undefined);
  assert.equal(page.win.lfLangLocked(), false);
  page.win.lfSetLang('zh');
  assert.deepEqual(page.writes, [['loopforge.lang', 'zh']]);
  assert.equal(page.reloads(), 1);
}

// The policy script loads before boot.js on both pages.
for (const html of ['index.html', 'login.html']) {
  const src = read(html);
  const policy = src.indexOf('<script src="/platform-language.js"></script>');
  assert(policy > 0 && policy < src.indexOf('<script src="/boot.js"></script>'), `${html} must load the language policy before boot.js`);
  assert(src.includes('href="/preferences.css'), `${html} must load preferences.css`);
}

// Both language switches are hidden while the language is locked.
const css = read('preferences.css');
assert.match(css, /:root\[data-lang-locked="true"\] #langToggle,\s*:root\[data-lang-locked="true"\] #loginLang\{display:none!important\}/);
