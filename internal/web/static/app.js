const API = {
  send: '/v1/send-code',
  verify: '/v1/verify-code'
};

const CODE_TTL_SECONDS = 150;

const LANGS = ['ru', 'en', 'uz'];
const LANG_KEY = 'lang';

let currentPhone = '';
let countdownTimer = null;
let currentLang = 'ru';

// Timer state is stored (not just rendered) so it can be re-translated on language switch.
let timerState = 'idle'; // 'idle' | 'running' | 'expired'
let timerRemaining = 0;

// Alert is stored as a render function so it can be re-translated on language switch.
let alertState = null; // { render: () => string, type: string }

const I18N = {
  ru: {
    page_title: 'Авторизация через Telegram Gateway — живая демонстрация',
    lang_title: 'Язык интерфейса',
    theme_title: 'Тёмная тема',
    step1_title: 'Авторизация',
    step1_desc: 'Введите номер телефона. Сервис отправит код через официальный шлюз Telegram, а при недоступности — автоматически через SMS.',
    phone_label: 'Номер телефона',
    send_btn: 'Получить код',
    sending_btn: 'Отправка...',
    step2_title: 'Введите код',
    step2_SentTo: 'Код отправлен на номер ',
    code_label: 'Код подтверждения',
    verify_btn: 'Подтвердить',
    verifying_btn: 'Проверка...',
    change_phone_btn: 'Изменить номер',
    step3_title: '✅ Номер подтверждён!',
    step3_desc: 'Проверка выполнена локально в Redis за ~1 мс без дополнительных сетевых запросов к внешним шлюзам.',
    test_again_btn: 'Протестировать снова',
    timer_prefix: 'Повторная отправка доступна через ',
    timer_suffix: ' сек.',
    timer_expired: 'Можно запросить новый код.',
    retry_in: 'Повторите через {n} сек.',
    try_later: 'Повторите позже.',
    err_generic: 'Ошибка соединения с сервером.',
    err_invalid_request: 'Некорректный запрос.',
    err_invalid_phone: 'Неверный номер телефона. Используйте формат +998901234567.',
    err_invalid_code: 'Неверный код подтверждения.',
    err_code_expired: 'Срок действия кода истёк. Запросите новый.',
    err_max_attempts: 'Превышено число попыток. Запросите новый код.',
    err_rate_limit: 'Слишком много запросов.',
    err_already_sent: 'Код уже отправлен на этот номер. Введите его или дождитесь окончания срока действия.',
    err_otp_not_found: 'Код не найден или срок его действия истёк. Запросите новый.',
    err_not_reachable: 'Не удалось доставить код на этот номер.',
    err_sender_unavailable: 'Сервис отправки временно недоступен. Попробуйте позже.',
    err_internal: 'Внутренняя ошибка сервера.'
  },
  en: {
    page_title: 'Telegram Gateway Auth — Live Demo',
    lang_title: 'Interface language',
    theme_title: 'Dark theme',
    step1_title: 'Sign In',
    step1_desc: 'Enter your phone number. The service delivers the code via the official Telegram Gateway first, with automatic SMS fallback.',
    phone_label: 'Phone Number',
    send_btn: 'Send Code',
    sending_btn: 'Sending...',
    step2_title: 'Enter Code',
    step2_SentTo: 'Verification code sent to ',
    code_label: 'Verification Code',
    verify_btn: 'Verify Code',
    verifying_btn: 'Verifying...',
    change_phone_btn: 'Change Phone Number',
    step3_title: '✅ Verified Successfully!',
    step3_desc: 'Validated locally in Redis in ~1 ms, with no extra requests to external gateways.',
    test_again_btn: 'Test Again',
    timer_prefix: 'Resend available in ',
    timer_suffix: 's',
    timer_expired: 'You can request a new code now.',
    retry_in: 'Try again in {n}s.',
    try_later: 'Please try again later.',
    err_generic: 'Server connection error.',
    err_invalid_request: 'Invalid request.',
    err_invalid_phone: 'Invalid phone number. Use the format +998901234567.',
    err_invalid_code: 'Invalid verification code.',
    err_code_expired: 'The code has expired. Request a new one.',
    err_max_attempts: 'Too many attempts. Request a new code.',
    err_rate_limit: 'Too many requests.',
    err_already_sent: 'A code was already sent to this number. Enter it or wait until it expires.',
    err_otp_not_found: 'Code not found or expired. Request a new one.',
    err_not_reachable: 'Could not deliver the code to this number.',
    err_sender_unavailable: 'The sending service is temporarily unavailable. Try again later.',
    err_internal: 'Internal server error.'
  },
  uz: {
    page_title: 'Telegram Gateway orqali autentifikatsiya — jonli demo',
    lang_title: 'Interfeys tili',
    theme_title: 'Qorongʻi mavzu',
    step1_title: 'Avtorizatsiya',
    step1_desc: 'Telefon raqamingizni kiriting. Xizmat kodni Telegramning rasmiy shlyuzi orqali yuboradi, agar Telegram mavjud boʻlmasa — avtomatik ravishda SMS orqali.',
    phone_label: 'Telefon raqami',
    send_btn: 'Kod olish',
    sending_btn: 'Yuborilmoqda...',
    step2_title: 'Kodni kiriting',
    step2_SentTo: 'Kod quyidagi raqamga yuborildi: ',
    code_label: 'Tasdiqlash kodi',
    verify_btn: 'Tasdiqlash',
    verifying_btn: 'Tekshirilmoqda...',
    change_phone_btn: 'Raqamni oʻzgartirish',
    step3_title: '✅ Raqam tasdiqlandi!',
    step3_desc: 'Tekshiruv Redis’da lokal tarzda, ~1 ms ichida, tashqi shlyuzlarga qoʻshimcha soʻrovlarsiz bajarildi.',
    test_again_btn: 'Qayta sinab koʻrish',
    timer_prefix: 'Qayta yuborish ',
    timer_suffix: ' soniyadan keyin mumkin.',
    timer_expired: 'Endi yangi kod soʻrashingiz mumkin.',
    retry_in: '{n} soniyadan keyin qayta urinib koʻring.',
    try_later: 'Keyinroq qayta urinib koʻring.',
    err_generic: 'Serverga ulanishda xatolik.',
    err_invalid_request: 'Notoʻgʻri soʻrov.',
    err_invalid_phone: 'Telefon raqami notoʻgʻri. +998901234567 formatidan foydalaning.',
    err_invalid_code: 'Tasdiqlash kodi notoʻgʻri.',
    err_code_expired: 'Kodning amal qilish muddati tugagan. Yangi kod soʻrang.',
    err_max_attempts: 'Urinishlar soni oshib ketdi. Yangi kod soʻrang.',
    err_rate_limit: 'Soʻrovlar juda koʻp.',
    err_already_sent: 'Bu raqamga kod allaqachon yuborilgan. Uni kiriting yoki amal qilish muddati tugashini kuting.',
    err_otp_not_found: 'Kod topilmadi yoki muddati tugagan. Yangi kod soʻrang.',
    err_not_reachable: 'Kodni bu raqamga yetkazib boʻlmadi.',
    err_sender_unavailable: 'Yuborish xizmati vaqtincha ishlamayapti. Keyinroq urinib koʻring.',
    err_internal: 'Serverning ichki xatosi.'
  }
};

// Backend error code (response "code" field) -> i18n key.
const ERROR_KEYS = {
  INVALID_REQUEST: 'err_invalid_request',
  INVALID_PHONE_NUMBER: 'err_invalid_phone',
  INVALID_CODE: 'err_invalid_code',
  CODE_EXPIRED: 'err_code_expired',
  MAX_ATTEMPTS_EXCEEDED: 'err_max_attempts',
  TOO_MANY_REQUESTS: 'err_rate_limit',
  MESSAGE_ALREADY_SENT: 'err_already_sent',
  OTP_NOT_FOUND: 'err_otp_not_found',
  PHONE_NOT_REACHABLE: 'err_not_reachable',
  SENDER_UNAVAILABLE: 'err_sender_unavailable',
  INTERNAL_ERROR: 'err_internal'
};

// Errors after which the current OTP session is over.
const SESSION_ENDED = ['MAX_ATTEMPTS_EXCEEDED', 'OTP_NOT_FOUND', 'CODE_EXPIRED'];

function t(key) {
  return I18N[currentLang][key];
}

// ---- Theme: saved choice, otherwise the system preference ----
const THEME_KEY = 'theme';
const systemDark = window.matchMedia('(prefers-color-scheme: dark)');

function savedTheme() {
  try {
    const v = localStorage.getItem(THEME_KEY);
    return v === 'light' || v === 'dark' ? v : null;
  } catch (_) {
    return null;
  }
}

function applyTheme(theme) {
  document.documentElement.setAttribute('data-theme', theme);
  const btn = document.getElementById('theme-btn');
  if (btn) btn.setAttribute('aria-pressed', String(theme === 'dark'));
}

function toggleTheme() {
  const next = document.documentElement.getAttribute('data-theme') === 'dark' ? 'light' : 'dark';
  applyTheme(next);
  try { localStorage.setItem(THEME_KEY, next); } catch (_) {}
}

// ---- Language: saved choice, otherwise the browser language, otherwise Russian ----
function detectLang() {
  try {
    const saved = localStorage.getItem(LANG_KEY);
    if (LANGS.includes(saved)) return saved;
  } catch (_) {}
  const prefs = navigator.languages && navigator.languages.length ? navigator.languages : [navigator.language || ''];
  for (const p of prefs) {
    const code = String(p).toLowerCase().slice(0, 2);
    if (LANGS.includes(code)) return code;
  }
  return 'ru';
}

function applyLanguage(lang) {
  currentLang = lang;
  document.documentElement.lang = lang;
  document.title = t('page_title');

  document.querySelectorAll('[data-i18n]').forEach((el) => {
    const text = I18N[lang][el.getAttribute('data-i18n')];
    if (text) el.textContent = text;
  });

  const langSwitch = document.getElementById('lang-switch');
  langSwitch.style.setProperty('--i', LANGS.indexOf(lang));
  langSwitch.setAttribute('aria-label', t('lang_title'));
  langSwitch.querySelectorAll('.lang-opt').forEach((opt) => {
    const active = opt.dataset.lang === lang;
    opt.setAttribute('aria-checked', String(active));
    opt.tabIndex = active ? 0 : -1; // roving tabindex: one tab stop for the whole group
  });

  const themeBtn = document.getElementById('theme-btn');
  themeBtn.title = t('theme_title');
  themeBtn.setAttribute('aria-label', t('theme_title'));

  if (!document.getElementById('step-verify').classList.contains('hidden')) {
    updateVerifySubtitle();
  }
  paintAlert();
  paintTimer();
}

function setLanguage(lang) {
  if (lang === currentLang) return;
  applyLanguage(lang);
  try { localStorage.setItem(LANG_KEY, lang); } catch (_) {}
}

// One click to any language; arrow keys / Home / End move between options (radio-group pattern).
function initLanguageSwitch() {
  const sw = document.getElementById('lang-switch');

  sw.addEventListener('click', (e) => {
    const opt = e.target.closest('.lang-opt');
    if (opt) setLanguage(opt.dataset.lang);
  });

  sw.addEventListener('keydown', (e) => {
    const step = { ArrowRight: 1, ArrowDown: 1, ArrowLeft: -1, ArrowUp: -1 }[e.key];
    let next = null;
    if (step) next = LANGS[(LANGS.indexOf(currentLang) + step + LANGS.length) % LANGS.length];
    else if (e.key === 'Home') next = LANGS[0];
    else if (e.key === 'End') next = LANGS[LANGS.length - 1];
    if (!next) return;

    e.preventDefault();
    setLanguage(next);
    sw.querySelector('[data-lang="' + next + '"]').focus();
  });

  // Enable the thumb animation only after the initial position is set, so it does not slide on page load.
  requestAnimationFrame(() => sw.classList.add('ready'));
}

// ---- Phone: formatting and validation ----
// Uses libphonenumber-js when /static/libphonenumber-min.js is loaded,
// otherwise falls back to a simple formatter and an E.164 regex.
function formatPhone(raw) {
  const digits = raw.replace(/\D/g, '').slice(0, 15); // E.164: up to 15 digits
  if (!digits) return '';

  const lib = window.libphonenumber;
  if (lib) return new lib.AsYouType().input('+' + digits);

  const groups = [3, 2, 3, 2, 2]; // 998 90 123 45 67
  const parts = [];
  let pos = 0;
  for (const size of groups) {
    if (pos >= digits.length) break;
    parts.push(digits.slice(pos, pos + size));
    pos += size;
  }
  if (pos < digits.length) parts.push(digits.slice(pos));
  return '+' + parts.join(' ');
}

// Returns the number in E.164 format or null if it is invalid.
function normalizePhone(raw) {
  const lib = window.libphonenumber;
  if (lib) {
    const parsed = lib.parsePhoneNumberFromString(raw);
    return parsed && parsed.isValid() ? parsed.number : null;
  }
  const e164 = '+' + raw.replace(/\D/g, '');
  return /^\+[1-9]\d{6,14}$/.test(e164) ? e164 : null;
}

// ---- Alerts ----
// `render` is a function returning the message, so the text follows the current language.
function paintAlert() {
  const box = document.getElementById('alert-box');
  if (!alertState) {
    box.style.display = 'none';
    return;
  }
  box.textContent = alertState.render();
  box.className = 'alert-box' + (alertState.type ? ' ' + alertState.type : '');
  box.style.display = 'block';
}

function showAlert(render, type) {
  alertState = { render, type };
  paintAlert();
}

function hideAlert() {
  alertState = null;
  paintAlert();
}

// ---- Network ----
async function postJSON(url, payload) {
  const res = await fetch(url, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload)
  });
  let data = {};
  try { data = await res.json(); } catch (_) {}
  return { res, data };
}

// Returns a render function: the message is built from the current language on every call.
function errorMessage(res, data) {
  const key = ERROR_KEYS[data.code];
  const isRateLimited = res.status === 429;
  const retryAfter = isRateLimited ? parseInt(res.headers.get('Retry-After'), 10) : NaN;

  return () => {
    let msg = key ? t(key) : t('err_generic');
    if (isRateLimited) {
      msg += ' ' + (retryAfter > 0 ? t('retry_in').replace('{n}', retryAfter) : t('try_later'));
    }
    return msg;
  };
}

// ---- Flow ----
async function handleSendOTP(e) {
  e.preventDefault();
  hideAlert();

  const phone = normalizePhone(document.getElementById('phone-input').value);
  if (!phone) {
    showAlert(() => t('err_invalid_phone'));
    return;
  }
  currentPhone = phone;

  const btn = document.getElementById('send-btn');
  btn.disabled = true;
  btn.textContent = t('sending_btn');

  try {
    const { res, data } = await postJSON(API.send, { phone: currentPhone });

    if (res.ok) {
      renderVerifyStep(true);
    } else if (data.code === 'MESSAGE_ALREADY_SENT') {
      // A code for this phone is still valid: let the user enter it.
      renderVerifyStep(false);
      showAlert(errorMessage(res, data), 'info');
    } else {
      showAlert(errorMessage(res, data));
    }
  } catch (err) {
    showAlert(() => t('err_generic'));
  } finally {
    btn.disabled = false;
    btn.textContent = t('send_btn');
  }
}

function updateVerifySubtitle() {
  document.getElementById('verify-subtitle').textContent = t('step2_SentTo') + formatPhone(currentPhone);
}

function renderVerifyStep(startTimer) {
  document.getElementById('step-phone').classList.add('hidden');
  document.getElementById('step-success').classList.add('hidden');
  document.getElementById('step-verify').classList.remove('hidden');

  updateVerifySubtitle();

  const codeInput = document.getElementById('code-input');
  codeInput.value = '';
  codeInput.focus();

  if (startTimer) {
    startCountdown(CODE_TTL_SECONDS);
  } else {
    stopCountdown();
  }
}

async function handleVerifyOTP(e) {
  e.preventDefault();
  hideAlert();

  const code = document.getElementById('code-input').value.trim();
  const btn = document.getElementById('verify-btn');
  btn.disabled = true;
  btn.textContent = t('verifying_btn');

  try {
    const { res, data } = await postJSON(API.verify, { phone: currentPhone, code });

    if (res.ok) {
      stopCountdown();
      document.getElementById('step-verify').classList.add('hidden');
      document.getElementById('step-success').classList.remove('hidden');
      return;
    }

    const msg = errorMessage(res, data);
    if (SESSION_ENDED.includes(data.code)) {
      resetToPhone();
      showAlert(msg);
      return;
    }

    showAlert(msg);
    document.getElementById('code-input').select();
  } catch (err) {
    showAlert(() => t('err_generic'));
  } finally {
    btn.disabled = false;
    btn.textContent = t('verify_btn');
  }
}

function resetToPhone() {
  hideAlert();
  stopCountdown();
  document.getElementById('step-verify').classList.add('hidden');
  document.getElementById('step-success').classList.add('hidden');
  document.getElementById('step-phone').classList.remove('hidden');
}

// ---- Countdown ----
function paintTimer() {
  const label = document.getElementById('timer-label');
  if (timerState === 'running') {
    label.textContent = t('timer_prefix') + timerRemaining + t('timer_suffix');
  } else if (timerState === 'expired') {
    label.textContent = t('timer_expired');
  } else {
    label.textContent = '';
  }
}

function stopCountdown() {
  if (countdownTimer) clearInterval(countdownTimer);
  countdownTimer = null;
  timerState = 'idle';
  paintTimer();
}

function startCountdown(seconds) {
  stopCountdown();
  timerRemaining = seconds;
  timerState = 'running';

  const tick = () => {
    if (timerRemaining <= 0) {
      clearInterval(countdownTimer);
      countdownTimer = null;
      timerState = 'expired';
      paintTimer();
      return;
    }
    paintTimer();
    timerRemaining--;
  };
  tick();
  countdownTimer = setInterval(tick, 1000);
}

document.addEventListener('DOMContentLoaded', () => {
  applyTheme(savedTheme() || (systemDark.matches ? 'dark' : 'light'));
  // Follow the system theme until the user picks one manually.
  systemDark.addEventListener('change', (e) => {
    if (!savedTheme()) applyTheme(e.matches ? 'dark' : 'light');
  });

  applyLanguage(detectLang());
  initLanguageSwitch();

  const phoneInput = document.getElementById('phone-input');
  phoneInput.addEventListener('input', (e) => {
    // Do not reformat while editing in the middle, so the caret does not jump.
    if (e.target.selectionStart !== e.target.value.length) return;
    e.target.value = formatPhone(e.target.value);
  });

  document.getElementById('form-send').addEventListener('submit', handleSendOTP);
  document.getElementById('form-verify').addEventListener('submit', handleVerifyOTP);
  document.getElementById('change-phone-btn').addEventListener('click', resetToPhone);
  document.getElementById('test-again-btn').addEventListener('click', resetToPhone);
  document.getElementById('theme-btn').addEventListener('click', toggleTheme);
});