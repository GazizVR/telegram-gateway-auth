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

// Verification timer state ('idle' | 'running' | 'expired')
let timerState = 'idle';
let timerRemaining = 0;

// Flood wait live timer state
let floodTimer = null;
let floodRemaining = 0;

// Alert state stored as a render function so it updates on language changes
let alertState = null; // { render: () => string, type: string }

// Storage for active OTP countdowns per phone number
const SESSIONS_STORAGE_KEY = 'otp_active_sessions';

function getSessionExpiresAt(phone) {
  if (!phone) return 0;
  try {
    const raw = sessionStorage.getItem(SESSIONS_STORAGE_KEY);
    if (!raw) return 0;
    const map = JSON.parse(raw);
    return typeof map[phone] === 'number' ? map[phone] : 0;
  } catch (_) {
    return 0;
  }
}

function setSessionExpiresAt(phone, expiresAt) {
  if (!phone) return;
  try {
    const raw = sessionStorage.getItem(SESSIONS_STORAGE_KEY);
    const map = raw ? JSON.parse(raw) : {};
    map[phone] = expiresAt;
    sessionStorage.setItem(SESSIONS_STORAGE_KEY, JSON.stringify(map));
  } catch (_) {}
}

function removeSession(phone) {
  if (!phone) return;
  try {
    const raw = sessionStorage.getItem(SESSIONS_STORAGE_KEY);
    if (!raw) return;
    const map = JSON.parse(raw);
    delete map[phone];
    sessionStorage.setItem(SESSIONS_STORAGE_KEY, JSON.stringify(map));
  } catch (_) {}
}

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
    err_flood_wait: 'Слишком много запросов (Flood Wait).',
    flood_wait_active: 'Слишком много запросов. Повторная отправка доступна через {n} сек.',
    flood_wait_expired: 'Время ожидания истекло. Вы можете запросить код снова.',
    flood_btn_wait: 'Подождите ({n}с)...',
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
    err_flood_wait: 'Too many requests (Flood Wait).',
    flood_wait_active: 'Too many requests. Resend available in {n}s.',
    flood_wait_expired: 'Wait time expired. You can request a code again.',
    flood_btn_wait: 'Wait ({n}s)...',
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
    err_flood_wait: 'Soʻrovlar juda koʻp (Flood Wait).',
    flood_wait_active: 'Soʻrovlar juda koʻp. Qayta yuborish {n} soniyadan keyin mumkin.',
    flood_wait_expired: 'Kutish vaqti tugadi. Endi qayta kod soʻrashingiz mumkin.',
    flood_btn_wait: 'Kuting ({n}s)...',
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
  FLOOD_WAIT: 'err_flood_wait',
  MESSAGE_ALREADY_SENT: 'err_already_sent',
  OTP_NOT_FOUND: 'err_otp_not_found',
  PHONE_NOT_REACHABLE: 'err_not_reachable',
  SENDER_UNAVAILABLE: 'err_sender_unavailable',
  INTERNAL_ERROR: 'err_internal'
};

// Errors after which the current OTP session is considered expired/cancelled
const SESSION_ENDED = ['MAX_ATTEMPTS_EXCEEDED', 'OTP_NOT_FOUND', 'CODE_EXPIRED'];

function t(key) {
  return (I18N[currentLang] && I18N[currentLang][key]) || (I18N.ru && I18N.ru[key]) || key;
}

// ---- Theme: saved choice, otherwise system preference ----
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

// ---- Language management ----
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
  if (langSwitch) {
    langSwitch.style.setProperty('--i', LANGS.indexOf(lang));
    langSwitch.setAttribute('aria-label', t('lang_title'));
    langSwitch.querySelectorAll('.lang-opt').forEach((opt) => {
      const active = opt.dataset.lang === lang;
      opt.setAttribute('aria-checked', String(active));
      opt.tabIndex = active ? 0 : -1;
    });
  }

  const themeBtn = document.getElementById('theme-btn');
  if (themeBtn) {
    themeBtn.title = t('theme_title');
    themeBtn.setAttribute('aria-label', t('theme_title'));
  }

  if (!document.getElementById('step-verify').classList.contains('hidden')) {
    updateVerifySubtitle();
  }

  // Update live flood wait UI button text if countdown is running
  if (floodRemaining > 0) {
    const sendBtn = document.getElementById('send-btn');
    if (sendBtn) {
      sendBtn.textContent = t('flood_btn_wait').replace('{n}', floodRemaining);
    }
  }

  paintAlert();
  paintTimer();
}

function setLanguage(lang) {
  if (lang === currentLang) return;
  applyLanguage(lang);
  try { localStorage.setItem(LANG_KEY, lang); } catch (_) {}
}

function initLanguageSwitch() {
  const sw = document.getElementById('lang-switch');
  if (!sw) return;

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

  requestAnimationFrame(() => sw.classList.add('ready'));
}

// ---- Phone formatting and validation ----
function formatPhone(raw) {
  const digits = raw.replace(/\D/g, '').slice(0, 15);
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

function normalizePhone(raw) {
  const lib = window.libphonenumber;
  if (lib) {
    const parsed = lib.parsePhoneNumberFromString(raw);
    return parsed && parsed.isValid() ? parsed.number : null;
  }
  const e164 = '+' + raw.replace(/\D/g, '');
  return /^\+[1-9]\d{6,14}$/.test(e164) ? e164 : null;
}

// ---- Alert Box ----
function paintAlert() {
  const box = document.getElementById('alert-box');
  if (!alertState) {
    box.style.display = 'none';
    box.textContent = '';
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

// ---- Live Flood Wait Countdown Timer ----
function extractRetrySeconds(res, data) {
  if (data && typeof data.retry_after === 'number' && data.retry_after > 0) {
    return data.retry_after;
  }
  const headerVal = res && res.headers ? parseInt(res.headers.get('Retry-After'), 10) : NaN;
  if (!isNaN(headerVal) && headerVal > 0) {
    return headerVal;
  }
  if (data && typeof data.message === 'string') {
    const m = data.message.match(/retry after (?:(\d+)m)?(?:(\d+)s)?/i);
    if (m) {
      const mins = parseInt(m[1] || '0', 10);
      const secs = parseInt(m[2] || '0', 10);
      const total = mins * 60 + secs;
      if (total > 0) return total;
    }
    const digits = data.message.match(/(\d+)\s*(?:s|sec|сек|soniya)/i);
    if (digits) {
      const val = parseInt(digits[1], 10);
      if (val > 0) return val;
    }
  }
  return 60; // fallback standard cooldown
}

function startFloodCountdown(seconds) {
  stopFloodCountdown(false);
  floodRemaining = seconds;

  const sendBtn = document.getElementById('send-btn');
  if (sendBtn) {
    sendBtn.disabled = true;
    sendBtn.textContent = t('flood_btn_wait').replace('{n}', floodRemaining);
  }

  showAlert(() => t('flood_wait_active').replace('{n}', floodRemaining), 'warning');

  floodTimer = setInterval(() => {
    floodRemaining--;
    if (floodRemaining <= 0) {
      stopFloodCountdown(true);
      showAlert(() => t('flood_wait_expired'), 'info');
      return;
    }

    if (sendBtn) {
      sendBtn.textContent = t('flood_btn_wait').replace('{n}', floodRemaining);
    }
    paintAlert();
  }, 1000);
}

function stopFloodCountdown(resetButton = true) {
  if (floodTimer) {
    clearInterval(floodTimer);
    floodTimer = null;
  }
  floodRemaining = 0;
  if (resetButton) {
    const sendBtn = document.getElementById('send-btn');
    if (sendBtn) {
      sendBtn.disabled = false;
      sendBtn.textContent = t('send_btn');
    }
  }
}

// ---- Network Helpers ----
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

function errorMessage(res, data) {
  const key = data && data.code ? ERROR_KEYS[data.code] : null;

  return () => {
    if (key) return t(key);
    if (data && typeof data.message === 'string' && data.message.trim()) {
      return data.message;
    }
    return t('err_generic');
  };
}

// ---- User Flow ----
async function handleSendOTP(e) {
  e.preventDefault();

  if (floodRemaining > 0) {
    showAlert(() => t('flood_wait_active').replace('{n}', floodRemaining), 'warning');
    return;
  }

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
    } else if (data.code === 'FLOOD_WAIT' || res.status === 429) {
      const waitSeconds = extractRetrySeconds(res, data);
      startFloodCountdown(waitSeconds);
    } else if (data.code === 'MESSAGE_ALREADY_SENT') {
      // Code is already issued and active: continue existing countdown!
      renderVerifyStep(false);
      showAlert(errorMessage(res, data), 'info');
    } else {
      showAlert(errorMessage(res, data));
    }
  } catch (err) {
    showAlert(() => t('err_generic'));
  } finally {
    // Keep button disabled if flood wait countdown is active
    if (floodRemaining <= 0) {
      btn.disabled = false;
      btn.textContent = t('send_btn');
    }
  }
}

function updateVerifySubtitle() {
  const el = document.getElementById('verify-subtitle');
  if (el) {
    el.textContent = t('step2_SentTo') + formatPhone(currentPhone);
  }
}

function renderVerifyStep(isNewCode = false) {
  document.getElementById('step-phone').classList.add('hidden');
  document.getElementById('step-success').classList.add('hidden');
  document.getElementById('step-verify').classList.remove('hidden');

  updateVerifySubtitle();

  const codeInput = document.getElementById('code-input');
  codeInput.value = '';
  codeInput.focus();

  startCountdownForPhone(currentPhone, isNewCode);
}

async function handleVerifyOTP(e) {
  e.preventDefault();
  hideAlert();

  const codeInput = document.getElementById('code-input');
  const code = codeInput.value.replace(/\D/g, '').trim();

  if (code.length === 0) {
    showAlert(() => t('err_invalid_code'));
    codeInput.focus();
    return;
  }

  const btn = document.getElementById('verify-btn');
  btn.disabled = true;
  btn.textContent = t('verifying_btn');

  try {
    const { res, data } = await postJSON(API.verify, { phone: currentPhone, code });

    if (res.ok) {
      removeSession(currentPhone);
      stopCountdown();
      document.getElementById('step-verify').classList.add('hidden');
      document.getElementById('step-success').classList.remove('hidden');
      return;
    }

    if (data.code === 'FLOOD_WAIT' || res.status === 429) {
      const waitSeconds = extractRetrySeconds(res, data);
      startFloodCountdown(waitSeconds);
    }

    const msg = errorMessage(res, data);
    if (SESSION_ENDED.includes(data.code)) {
      removeSession(currentPhone);
      resetToPhone();
      showAlert(msg);
      return;
    }

    showAlert(msg);
    codeInput.select();
  } catch (err) {
    showAlert(() => t('err_generic'));
  } finally {
    btn.disabled = false;
    btn.textContent = t('verify_btn');
  }
}

function resetToPhone() {
  hideAlert();
  // Clear the interval tick, but preserve session expiration timestamp in storage
  if (countdownTimer) {
    clearInterval(countdownTimer);
    countdownTimer = null;
  }
  timerState = 'idle';

  document.getElementById('step-verify').classList.add('hidden');
  document.getElementById('step-success').classList.add('hidden');
  document.getElementById('step-phone').classList.remove('hidden');

  const phoneInput = document.getElementById('phone-input');
  if (phoneInput) {
    phoneInput.focus();
  }
}

function handleTestAgain() {
  removeSession(currentPhone);
  currentPhone = '';
  const phoneInput = document.getElementById('phone-input');
  if (phoneInput) {
    phoneInput.value = '';
  }
  const codeInput = document.getElementById('code-input');
  if (codeInput) {
    codeInput.value = '';
  }
  stopFloodCountdown(true);
  resetToPhone();
}

// ---- Verification Countdown ----
function paintTimer() {
  const label = document.getElementById('timer-label');
  if (!label) return;

  if (timerState === 'running') {
    label.textContent = t('timer_prefix') + timerRemaining + t('timer_suffix');
  } else if (timerState === 'expired') {
    label.textContent = t('timer_expired');
  } else {
    label.textContent = '';
  }
}

function stopCountdown() {
  if (countdownTimer) {
    clearInterval(countdownTimer);
    countdownTimer = null;
  }
  timerState = 'idle';
  timerRemaining = 0;
  paintTimer();
}

function startCountdownForPhone(phone, isNewCode = false) {
  let expiresAt = getSessionExpiresAt(phone);
  const now = Date.now();

  if (isNewCode || !expiresAt || expiresAt <= now) {
    expiresAt = now + CODE_TTL_SECONDS * 1000;
    setSessionExpiresAt(phone, expiresAt);
  }

  runCountdown(expiresAt);
}

function runCountdown(expiresAt) {
  if (countdownTimer) {
    clearInterval(countdownTimer);
    countdownTimer = null;
  }

  const tick = () => {
    const remainingMs = expiresAt - Date.now();
    timerRemaining = Math.max(0, Math.ceil(remainingMs / 1000));

    if (timerRemaining <= 0) {
      if (countdownTimer) {
        clearInterval(countdownTimer);
        countdownTimer = null;
      }
      timerState = 'expired';
      removeSession(currentPhone);
      paintTimer();
      return;
    }

    timerState = 'running';
    paintTimer();
  };

  tick();
  countdownTimer = setInterval(tick, 1000);
}

// ---- Initialization ----
document.addEventListener('DOMContentLoaded', () => {
  applyTheme(savedTheme() || (systemDark.matches ? 'dark' : 'light'));
  systemDark.addEventListener('change', () => {
    if (!savedTheme()) applyTheme(systemDark.matches ? 'dark' : 'light');
  });

  applyLanguage(detectLang());
  initLanguageSwitch();

  const phoneInput = document.getElementById('phone-input');
  if (phoneInput) {
    phoneInput.addEventListener('input', (e) => {
      if (e.target.selectionStart !== e.target.value.length) return;
      e.target.value = formatPhone(e.target.value);
    });
  }

  const codeInput = document.getElementById('code-input');
  if (codeInput) {
    codeInput.addEventListener('input', (e) => {
      const clean = e.target.value.replace(/\D/g, '').slice(0, 6);
      e.target.value = clean;
      // Auto-submit verification once all 6 digits are typed
      if (clean.length === 6) {
        const formVerify = document.getElementById('form-verify');
        if (formVerify && typeof formVerify.requestSubmit === 'function') {
          formVerify.requestSubmit();
        }
      }
    });
  }

  document.getElementById('form-send').addEventListener('submit', handleSendOTP);
  document.getElementById('form-verify').addEventListener('submit', handleVerifyOTP);
  document.getElementById('change-phone-btn').addEventListener('click', resetToPhone);
  document.getElementById('test-again-btn').addEventListener('click', handleTestAgain);
  document.getElementById('theme-btn').addEventListener('click', toggleTheme);
});