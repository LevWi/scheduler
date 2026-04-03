(() => {
  const tg = window.Telegram?.WebApp;
  const params = new URLSearchParams(location.search);
  tg?.ready();
  tg?.expand();

  const state = {
    apiBase: params.get('api_base') || '',
    clientId: params.get('telegram_bot_id')
      || params.get('bot_id')
      || '',
    monthDate: new Date(),
    selectedDate: null,
    selectedSlot: null,
    monthSlots: [],
    daySlots: [],
    isSubmitting: false,
  };

  const quickSlotsEl = document.getElementById('quick-slots');
  const monthLabelEl = document.getElementById('month-label');
  const calendarGridEl = document.getElementById('calendar-grid');
  const daySlotsEl = document.getElementById('day-slots');
  const slotsTitleEl = document.getElementById('slots-title');
  const confirmBtn = document.getElementById('confirm-btn');

  const ruMonth = new Intl.DateTimeFormat('ru-RU', { month: 'long', year: 'numeric' });
  const ruDow = new Intl.DateTimeFormat('ru-RU', { weekday: 'short' });
  const ruDate = new Intl.DateTimeFormat('ru-RU', { day: '2-digit', month: '2-digit' });
  const ruTime = new Intl.DateTimeFormat('ru-RU', { hour: '2-digit', minute: '2-digit' });

  applyTelegramTheme();
  bindEvents();
  bootstrap().catch((err) => renderMessage(daySlotsEl, `Ошибка загрузки: ${err.message}`));

  async function bootstrap() {
    if (!state.clientId) {
      renderMessage(quickSlotsEl, 'Передайте telegram_bot_id или bot_id в query string');
      return;
    }
    if (!tg?.initData) {
      renderMessage(quickSlotsEl, 'Telegram initData отсутствует');
      return;
    }
    await Promise.all([loadQuickSlots(), loadMonth(state.monthDate)]);
  }

  function bindEvents() {
    document.getElementById('prev-month').addEventListener('click', () => shiftMonth(-1));
    document.getElementById('next-month').addEventListener('click', () => shiftMonth(1));
    confirmBtn.addEventListener('click', () => submitSelection());
  }

  async function shiftMonth(offset) {
    state.monthDate = new Date(state.monthDate.getFullYear(), state.monthDate.getMonth() + offset, 1);
    await loadMonth(state.monthDate);
  }

  async function loadQuickSlots() {
    const now = new Date();
    const end = new Date(now);
    end.setDate(end.getDate() + 3);
    const resp = await getSlots(now, end);
    const byDay = new Map();
    for (const slot of resp.slots || []) {
      const d = new Date(slot.tp_start);
      const key = dateKey(d);
      if (!byDay.has(key)) byDay.set(key, slot);
    }

    quickSlotsEl.innerHTML = '';
    for (let i = 0; i < 3; i++) {
      const day = new Date();
      day.setHours(0, 0, 0, 0);
      day.setDate(day.getDate() + i);
      const key = dateKey(day);
      const slot = byDay.get(key);
      const btn = document.createElement('button');
      btn.className = 'quick-slot';
      if (slot) {
        const start = new Date(slot.tp_start);
        btn.innerHTML = `<strong>${capitalize(ruDow.format(start))}, ${ruDate.format(start)}</strong><br><small>${ruTime.format(start)}</small>`;
      } else {
        btn.innerHTML = `<strong>${capitalize(ruDow.format(day))}, ${ruDate.format(day)}</strong><br><small>Нет слотов</small>`;
      }
      btn.addEventListener('click', async () => {
        [...quickSlotsEl.children].forEach((el) => el.classList.remove('active'));
        btn.classList.add('active');
        state.selectedDate = key;
        state.selectedSlot = null;
        await loadDaySlots(day);
        renderCalendar();
      });
      quickSlotsEl.appendChild(btn);
    }
  }

  async function loadMonth(date) {
    calendarGridEl.classList.add('loading');
    const monthStart = new Date(date.getFullYear(), date.getMonth(), 1);
    const monthEnd = new Date(date.getFullYear(), date.getMonth() + 1, 1);
    const resp = await getSlots(monthStart, monthEnd);
    state.monthSlots = resp.slots || [];
    monthLabelEl.textContent = capitalize(ruMonth.format(monthStart));
    renderCalendar();
    calendarGridEl.classList.remove('loading');
  }

  function renderCalendar() {
    calendarGridEl.innerHTML = '';
    const monthStart = new Date(state.monthDate.getFullYear(), state.monthDate.getMonth(), 1);
    const firstDay = (monthStart.getDay() + 6) % 7;
    const gridStart = new Date(monthStart);
    gridStart.setDate(monthStart.getDate() - firstDay);

    const daysWithSlots = new Set(state.monthSlots.map((slot) => dateKey(new Date(slot.tp_start))));
    const todayKey = dateKey(new Date());

    for (let i = 0; i < 42; i++) {
      const d = new Date(gridStart);
      d.setDate(gridStart.getDate() + i);
      const key = dateKey(d);
      const btn = document.createElement('button');
      btn.className = 'day-btn';
      btn.textContent = String(d.getDate());

      if (d.getMonth() !== state.monthDate.getMonth()) btn.classList.add('other');
      if (daysWithSlots.has(key)) btn.classList.add('has-slots');
      if (key === todayKey) btn.classList.add('today');
      if (key === state.selectedDate) btn.classList.add('selected');

      btn.addEventListener('click', async () => {
        state.selectedDate = key;
        state.selectedSlot = null;
        await loadDaySlots(d);
        renderCalendar();
      });
      calendarGridEl.appendChild(btn);
    }
  }

  async function loadDaySlots(day) {
    const start = new Date(day);
    start.setHours(0, 0, 0, 0);
    const end = new Date(start);
    end.setDate(end.getDate() + 1);

    const resp = await getSlots(start, end);
    state.daySlots = resp.slots || [];
    slotsTitleEl.textContent = `Слоты на ${ruDate.format(start)}`;
    renderDaySlots();
  }

  function renderDaySlots() {
    daySlotsEl.innerHTML = '';
    setConfirmDisabled(true);

    if (!state.daySlots.length) {
      renderMessage(daySlotsEl, 'На выбранный день доступных слотов нет');
      return;
    }

    const now = Date.now();
    for (const slot of state.daySlots) {
      const start = new Date(slot.tp_start);
      const isBusy = start.getTime() < now;
      const btn = document.createElement('button');
      btn.className = 'slot-btn';
      btn.disabled = isBusy;
      btn.textContent = `${ruTime.format(start)} (${slot.len} мин)`;

      if (isBusy) btn.title = 'Слот уже недоступен';

      btn.addEventListener('click', () => {
        if (isBusy) return;
        state.selectedSlot = slot;
        [...daySlotsEl.children].forEach((el) => el.classList.remove('selected'));
        btn.classList.add('selected');
        setConfirmDisabled(false);
      });
      daySlotsEl.appendChild(btn);
    }
  }

  async function submitSelection() {
    if (!state.selectedSlot || state.isSubmitting) return;
    if (!state.clientId) {
      renderMessage(daySlotsEl, 'Передайте telegram_bot_id или bot_id в query string');
      return;
    }
    if (!tg?.initData) {
      renderMessage(daySlotsEl, 'Telegram initData отсутствует');
      return;
    }

    state.isSubmitting = true;
    setConfirmDisabled(true, 'Подтверждаем...');

    try {
      const url = `${state.apiBase}/slots/webapp`;
      const res = await fetch(url, {
        method: 'POST',
        headers: {
          Accept: 'application/json',
          'Content-Type': 'application/json; charset=UTF-8',
          Authorization: `tma ${tg.initData}`,
          'X-Client-ID': state.clientId,
        },
        body: JSON.stringify([state.selectedSlot]),
      });

      if (!res.ok) throw new Error(`HTTP ${res.status}`);

      tg?.HapticFeedback?.notificationOccurred('success');
      renderMessage(daySlotsEl, 'Запись подтверждена');
      setConfirmDisabled(true, 'Запись подтверждена');
      tg?.close();
    } catch (err) {
      tg?.HapticFeedback?.notificationOccurred('error');
      renderMessage(daySlotsEl, `Не удалось подтвердить запись: ${err.message}`);
      setConfirmDisabled(false);
    } finally {
      state.isSubmitting = false;
    }
  }

  async function getSlots(dateStart, dateEnd) {
    const qs = new URLSearchParams({
      date_start: dateStart.toISOString(),
      date_end: dateEnd.toISOString(),
    });
    const url = `${state.apiBase}/slots/webapp?${qs.toString()}`;
    const res = await fetch(url, {
      headers: {
        Accept: 'application/json',
        Authorization: `tma ${tg.initData}`,
        'X-Client-ID': state.clientId,
      },
    });
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    return res.json();
  }

  function dateKey(d) {
    const x = new Date(d);
    x.setHours(0, 0, 0, 0);
    return x.toISOString().slice(0, 10);
  }

  function renderMessage(root, text) {
    root.innerHTML = '';
    const p = document.createElement('p');
    p.className = 'msg';
    p.textContent = text;
    root.appendChild(p);
  }

  function setConfirmDisabled(disabled, text = 'Подтвердить запись') {
    confirmBtn.hidden = !state.selectedSlot && disabled && text === 'Подтвердить запись';
    confirmBtn.disabled = disabled;
    confirmBtn.textContent = text;
  }

  function capitalize(str) {
    return str.charAt(0).toUpperCase() + str.slice(1);
  }

  function applyTelegramTheme() {
    if (!tg?.themeParams) return;
    const tp = tg.themeParams;
    const root = document.documentElement;
    if (tp.bg_color) root.style.setProperty('--bg', tp.bg_color);
    if (tp.text_color) root.style.setProperty('--text', tp.text_color);
    if (tp.hint_color) root.style.setProperty('--hint', tp.hint_color);
    if (tp.secondary_bg_color) root.style.setProperty('--card', tp.secondary_bg_color);
    if (tp.button_color) root.style.setProperty('--accent', tp.button_color);
    if (tp.button_text_color) root.style.setProperty('--accent-contrast', tp.button_text_color);
  }
})();
