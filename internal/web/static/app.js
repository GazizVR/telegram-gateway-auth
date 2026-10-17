const API = {
  send: '/v1/send-code',
  verify: '/v1/verify-code'
};

// Matches otpService TTL on the backend (150s).
const CODE_TTL_SECONDS = 150;

let currentPhone = '';
let countdownTimer = null;
let currentLang = 'ru';

const I18N = {
  ru: {
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
    roi_title: 'Экономика для бизнеса',
    roi_badge: 'Экономия до 70%',
    sms_label: 'Обычный SMS-шлюз',
    timer_prefix: 'Повторная отправка доступна через ',
    timer_sec: ' сек.',
    timer_expired: 'Можно запросить новый код.',
    err_generic: 'Ошибка соединения с сервером.',
    err_invalid_request: 'Некорректный запрос.',
    err_invalid_phone: 'Неверный номер телефона. Используйте формат +998901234567.',
    err_invalid_code: 'Неверный код подтверждения.',
    err_code_expired: 'Срок действия кода истёк. Запросите новый.',
    err_max_attempts: 'Превышено число попыток. Запросите новый код.',
    err_rate_limit: 'Слишком много запросов. Повторите позже',
    err_already_sent: 'Код уже отправлен на этот номер. Введите его или дождитесь окончания срока действия.',
    err_otp_not_found: 'Код не найден или срок его действия истёк. Запросите новый.',
    err_not_reachable: 'Не удалось доставить код на этот номер.',
    err_sender_unavailable: 'Сервис отправки временно недоступен. Попробуйте позже.',
    err_internal: 'Внутренняя ошибка сервера.'
  },
  en: {
    step1_title: 'Sign In',
    step1_desc: 'Enter your phone number. The service delivers OTP via Telegram Gateway first, with automatic SMS fallback.',
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
    step3_desc: 'Validated locally in Redis in ~1ms with zero external API latency.',
    test_again_btn: 'Test Again',
    roi_title: 'Business Unit Economics',
    roi_badge: 'Save up to 70%',
    sms_label: 'Standard SMS Gateway',
    timer_prefix: 'Resend available in ',
    timer_sec: 's',
    timer_expired: 'You can request a new code now.',
    err_generic: 'Server connection error.',
    err_invalid_request: 'Invalid request.',
    err_invalid_phone: 'Invalid phone number. Use the format +998901234567.',
    err_invalid_code: 'Invalid verification code.',
    err_code_expired: 'The code has expired. Request a new one.',
    err_max_attempts: 'Too many attempts. Request a new code.',
    err_rate_limit: 'Too many requests. Please try again later',
    err_already_sent: 'A code was already sent to this number. Enter it or wait until it expires.',
    err_otp_not_found: 'Code not found or expired. Request a new one.',
    err_not_reachable: 'Could not deliver the code to this number.',
    err_sender_unavailable: 'Sending service is temporarily unavailable. Try again later.',
    err_internal: 'Internal server error.'
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

function showAlert(msg, type) {
  const box = document.getElementById('alert-box');
  box.textContent = msg;
  box.className = 'alert-box' + (type ? ' ' + type : '');
  box.style.display = 'block';
}

function hideAlert() {
  document.getElementById('alert-box').style.display = 'none';
}

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
  const key = ERROR_KEYS[data.code];
  let msg = key ? t(key) : t('err_generic');
  if (res.status === 429) {
    const retryAfter = parseInt(res.headers.get('Retry-After'), 10);
    msg += retryAfter > 0 ? ' (' + retryAfter + t('timer_sec') + ')' : '.';
  }
  return msg;
}

async function handleSendOTP(e) {
  e.preventDefault();
  hideAlert();

  currentPhone = document.getElementById('phone-input').value.replace(/\s+/g, '');

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
    showAlert(t('err_generic'));
  } finally {
    btn.disabled = false;
    btn.textContent = t('send_btn');
  }
}

function updateVerifySubtitle() {
  document.getElementById('verify-subtitle').textContent = t('step2_SentTo') + currentPhone;
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
    const codeInput = document.getElementById('code-input');
    codeInput.select();
  } catch (err) {
    showAlert(t('err_generic'));
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

function stopCountdown() {
  if (countdownTimer) clearInterval(countdownTimer);
  countdownTimer = null;
  document.getElementById('timer-label').textContent = '';
}

function startCountdown(seconds) {
  stopCountdown();
  let remaining = seconds;
  const label = document.getElementById('timer-label');

  const update = () => {
    if (remaining <= 0) {
      clearInterval(countdownTimer);
      countdownTimer = null;
      label.textContent = t('timer_expired');
      return;
    }
    label.textContent = t('timer_prefix') + remaining + t('timer_sec');
    remaining--;
  };
  update();
  countdownTimer = setInterval(update, 1000);
}

function toggleLanguage() {
  currentLang = currentLang === 'ru' ? 'en' : 'ru';
  document.getElementById('lang-btn').textContent = currentLang === 'ru' ? 'EN' : 'RU';

  document.querySelectorAll('[data-i18n]').forEach(el => {
    const key = el.getAttribute('data-i18n');
    if (I18N[currentLang][key]) {
      el.textContent = I18N[currentLang][key];
    }
  });

  if (!document.getElementById('step-verify').classList.contains('hidden')) {
    updateVerifySubtitle();
  }
}

function toggleTheme() {
  const html = document.documentElement;
  const next = html.getAttribute('data-theme') === 'dark' ? 'light' : 'dark';
  html.setAttribute('data-theme', next);
  document.getElementById('theme-btn').textContent = next === 'dark' ? '☀️' : '🌙';
}

document.addEventListener('DOMContentLoaded', () => {
  document.getElementById('form-send').addEventListener('submit', handleSendOTP);
  document.getElementById('form-verify').addEventListener('submit', handleVerifyOTP);
  document.getElementById('change-phone-btn').addEventListener('click', resetToPhone);
  document.getElementById('test-again-btn').addEventListener('click', resetToPhone);
  document.getElementById('lang-btn').addEventListener('click', toggleLanguage);
  document.getElementById('theme-btn').addEventListener('click', toggleTheme);
});