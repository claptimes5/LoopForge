// Apply the language before paint; boot-theme.js owns the theme.
// i18n.js replaces this. Boot must still render if that script fails to load.
window.lfT = window.lfT || ((zh) => zh);
(() => {
  const root = document.documentElement;
  // /platform-language.js sets LF_LANG_POLICY when a platform administrator
  // locks the interface language; the language switch is then hidden.
  const policy = window.LF_LANG_POLICY === 'en' || window.LF_LANG_POLICY === 'zh' ? window.LF_LANG_POLICY : '';
  let lang = 'zh';
  if (policy) {
    lang = policy;
    root.dataset.langLocked = 'true';
  } else {
    try {
      lang = localStorage.getItem('loopforge.lang') === 'en' ? 'en' : 'zh';
    } catch (_) { /* Restricted storage still permits the default interface. */ }
  }
  root.dataset.lang = lang;
  root.lang = lang === 'en' ? 'en' : 'zh-CN';
})();
