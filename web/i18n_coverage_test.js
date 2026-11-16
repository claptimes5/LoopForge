// English mode translates the Chinese UI at runtime through the dictionary in
// i18n-en.js. This test collects every Chinese string literal the main pages
// can render and runs it through the real translator; any string that still
// contains Chinese afterwards is missing from the dictionary.
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');

const WEB = __dirname;
// Files whose Chinese text is never rendered (icon animation notes).
const EXCLUDED_FILES = new Set(['app/harness-motion.js']);
// Strings that intentionally stay as-is in English mode.
// Strings that intentionally stay as-is in English mode: the language toggles,
// plus text that is never rendered (AI prompts, request payloads, the MCP setup
// prompt users copy to their agent, and regex sources that match user input).
const ALLOWED = new Set([
  '中 / EN', '中文 / English',
  '按照这个流程安装、初始化并验证好 LoopForge MCP：', '1. 使用 curl 下载 安装流程文件, 禁止使用 fetch：', '2. 两个占位符替换为：',
  '人工补充上下文：',
  '当前项目可用角色（agentId：名称）：', '当前画布上的 SOP：',
  'edges 用 {"source","target","when":"success"} 按成功路径连线，可以并行分支再汇合。',
  '去掉|移除|删除|取消|不用|不要', '静态站点|前端预览|预览节点|af-preview-build',
  // Date picker weekday headers switch language in hui-date.js itself.
  '一', '二', '三', '四', '五', '六', '日',
]);
// Translations that keep Chinese on purpose: language toggles, and a command
// the user must type to the WeCom bot verbatim.
const CHINESE_IN_ENGLISH_OK = /^(?:中文 \/ English|中文 ⇄ English|中 \/ EN)$|“绑定账号”|“开启新会话”/;

global.window = {confirm() {}, alert() {}, prompt() {}};
global.localStorage = {getItem() {}, setItem() {}};
global.MutationObserver = class { observe() {} };
global.document = {documentElement: {dataset: {lang: 'en'}}, body: null, addEventListener() {}};
eval(fs.readFileSync(path.join(WEB, 'i18n-en.js'), 'utf8'));
eval(fs.readFileSync(path.join(WEB, 'i18n.js'), 'utf8'));
const translate = window.lfTranslate;
const HAN = /[㐀-鿿]/;

for (const [zh, en] of Object.entries(window.LF_I18N_EN)) {
  assert(!HAN.test(en) || CHINESE_IN_ENGLISH_OK.test(en),`translation of ${JSON.stringify(zh)} still contains Chinese: ${JSON.stringify(en)}`);
}

const unescape = c => ({n: "\n", t: "\t", r: "\r"})[c] ?? c;

// String and template literal contents of a script, skipping comments and
// regex literals. Template literals yield the text between interpolations.
function jsStrings(src) {
  const out = [];
  const n = src.length;
  let i = 0;
  let prev = '';
  const regexAllowed = () => /^$|[(,=:[!&|?{};+\-*%<>~^]$|\b(?:return|typeof|case)$/.test(prev);
  function template() {
    let cur = '';
    while (i < n) {
      const c = src[i];
      if (c === '\\') { cur += unescape(src[i + 1]); i += 2; continue; }
      if (c === '`') { i++; out.push(cur); return; }
      if (c === '$' && src[i + 1] === '{') { out.push(cur); cur = ''; i += 2; code(true); continue; }
      cur += c; i++;
    }
    out.push(cur);
  }
  function code(untilBrace) {
    let depth = 0;
    while (i < n) {
      const c = src[i];
      if (untilBrace && c === '}' && depth === 0) { i++; return; }
      if (c === '{') depth++;
      if (c === '}') depth--;
      if (c === '/' && src[i + 1] === '/') { while (i < n && src[i] !== '\n') i++; continue; }
      if (c === '/' && src[i + 1] === '*') { const end = src.indexOf('*/', i + 2); i = end < 0 ? n : end + 2; continue; }
      if (c === "'" || c === '"') {
        let s = '';
        i++;
        while (i < n && src[i] !== c && src[i] !== '\n') {
          if (src[i] === '\\') { s += unescape(src[i + 1]); i += 2; } else s += src[i++];
        }
        i++; out.push(s); prev = 's'; continue;
      }
      if (c === '`') { i++; template(); prev = 's'; continue; }
      if (c === '/' && regexAllowed()) {
        i++;
        let inClass = false;
        while (i < n) {
          const d = src[i];
          if (d === '\\') { i += 2; continue; }
          if (d === '[') inClass = true;
          else if (d === ']') inClass = false;
          else if ((d === '/' && !inClass) || d === '\n') break;
          i++;
        }
        i++;
        while (/[a-z]/.test(src[i] || '')) i++;
        prev = 'r'; continue;
      }
      if (!/\s/.test(c)) prev = /[\w$]/.test(c) ? prev.match(/[\w$]*$/)[0] + c : c;
      i++;
    }
  }
  code(false);
  return out;
}

function htmlStrings(src) {
  src = src.replace(/<script[\s\S]*?<\/script>|<style[\s\S]*?<\/style>|<!--[\s\S]*?-->/g, '');
  const out = [];
  for (const m of src.matchAll(/>([^<]+)</g)) out.push(m[1]);
  for (const m of src.matchAll(/(?:placeholder|title|aria-label|data-tip|data-tooltip|value)="([^"]*)"/g)) out.push(m[1]);
  return out;
}

// Pages that load the translator, plus every local script they include.
const files = [];
for (const page of ['index.html', 'login.html']) {
  const html = fs.readFileSync(path.join(WEB, page), 'utf8');
  files.push(page);
  for (const m of html.matchAll(/<script[^>]*\ssrc="\/([^"?]+)/g)) {
    const file = m[1];
    if (file.startsWith('i18n') || EXCLUDED_FILES.has(file) || files.includes(file)) continue;
    if (fs.existsSync(path.join(WEB, file))) files.push(file);
  }
}
assert(files.length > 20, 'expected the app scripts to be discovered from index.html');

const missing = new Map();
for (const file of files) {
  const src = fs.readFileSync(path.join(WEB, file), 'utf8');
  const raw = file.endsWith('.html') ? htmlStrings(src) : jsStrings(src);
  for (const literal of raw) {
    for (const piece of literal.split(/<[^>]*>|&[a-z#0-9]+;|\n/)) {
      const text = piece.trim();
      const english = translate(text);
      if (!HAN.test(text) || ALLOWED.has(text) || !HAN.test(english) || CHINESE_IN_ENGLISH_OK.test(english)) continue;
      if (!missing.has(text)) missing.set(text, file);
    }
  }
}

if (missing.size) {
  const list = [...missing].map(([text, file]) => `  ${file}: ${JSON.stringify(text)}`).join('\n');
  assert.fail(`${missing.size} UI strings have no English translation in web/i18n-en.js:\n${list}`);
}
