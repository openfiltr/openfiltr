(function () {
  const REVIEW_STATES = [
    'validation-error',
    'error',
    'low-data',
    'add-record',
    'import-preview',
    'record-added',
    'add-rule',
    'rule-added',
    'empty',
  ];

  function cloneData(value) {
    if (value == null) {
      return value;
    }

    return JSON.parse(JSON.stringify(value));
  }

  function getQueryParam(key, search = window.location.search) {
    const params = new URLSearchParams(search);
    return params.get(key) || '';
  }

  function getReviewState(search = window.location.search) {
    const state = getQueryParam('state', search);

    if (!state || !REVIEW_STATES.includes(state)) {
      return 'default';
    }

    return state;
  }

  function toggleClass(target, className, shouldAdd) {
    if (!target) {
      return;
    }

    target.classList.toggle(className, Boolean(shouldAdd));
  }

  function setHidden(target, shouldHide) {
    if (!target) {
      return;
    }

    target.hidden = Boolean(shouldHide);
  }

  function setReviewStateClass(target, state, prefix = 'is-state-') {
    if (!target) {
      return;
    }

    REVIEW_STATES.forEach((reviewState) => {
      target.classList.remove(`${prefix}${reviewState}`);
    });

    if (state !== 'default') {
      target.classList.add(`${prefix}${state}`);
    }
  }

  function bootPage(initialise) {
    if (typeof initialise !== 'function') {
      return;
    }

    document.addEventListener('DOMContentLoaded', () => {
      initialise({
        state: getReviewState(),
        getQueryParam,
        toggleClass,
        setHidden,
        setReviewStateClass,
      });
    });
  }

  function validateSetupFields({ username = '', password = '', confirmPassword = '' }) {
    const trimmedUsername = username.trim();
    const result = {
      isValid: true,
      summary: 'Fix the highlighted fields to continue.',
      errors: {
        username: '',
        password: '',
        confirmPassword: '',
      },
    };

    if (!trimmedUsername) {
      result.isValid = false;
      result.errors.username = 'Enter a username.';
    }

    if (!password) {
      result.isValid = false;
      result.errors.password = 'Enter a password.';
    } else if (password.length < 8) {
      result.isValid = false;
      result.errors.password = 'Password must be at least 8 characters.';
    }

    if (!confirmPassword) {
      result.isValid = false;
      result.errors.confirmPassword = 'Confirm the password.';
    } else if (password && confirmPassword !== password) {
      result.isValid = false;
      result.errors.confirmPassword = 'Passwords do not match.';
    }

    const firstError = result.errors.username || result.errors.password || result.errors.confirmPassword;
    if (firstError) {
      result.summary = firstError;
    }

    return result;
  }

  function resolveLoginAttempt({ forceError = false, username = '', password = '' }) {
    if (forceError || !username.trim() || !password) {
      return { status: 'error' };
    }

    return { status: 'success' };
  }

  function resolveDashboardData(state = 'default') {
    const dashboard = window.OpenFiltrMockData && window.OpenFiltrMockData.dashboard;
    if (!dashboard) {
      return null;
    }

    const lowDataState = dashboard.reviewStates ? dashboard.reviewStates.lowData : 'low-data';
    if (state === lowDataState) {
      return cloneData(dashboard.lowData);
    }

    return cloneData(dashboard.normal);
  }

  function resolveActivityData(state = 'default', filter = '') {
    const activity = window.OpenFiltrMockData && window.OpenFiltrMockData.activity;
    if (!activity) {
      return null;
    }

    const base = state === 'low-data' ? activity.lowData : activity.normal;
    const data = cloneData(base);
    const normalisedFilter = filter || 'all';

    data.activeFilter = normalisedFilter;

    if (normalisedFilter === 'blocked') {
      data.items = data.items.filter((item) => item.action === 'Blocked');
    } else if (normalisedFilter === 'allowed') {
      data.items = data.items.filter((item) => item.action === 'Allowed');
    }

    return data;
  }

  function resolveDnsRecordsData(state = 'default') {
    const dnsRecords = window.OpenFiltrMockData && window.OpenFiltrMockData.dnsRecords;
    if (!dnsRecords) {
      return null;
    }

    const stateMap = {
      default: 'normal',
      empty: 'empty',
      'add-record': 'addRecord',
      'validation-error': 'validationError',
      'record-added': 'recordAdded',
      'import-preview': 'importPreview',
    };

    const recordStateKey = stateMap[state] || 'normal';
    const base = cloneData(state === 'empty' ? dnsRecords.empty : dnsRecords.normal);
    const derivedState = recordStateKey === 'normal' || recordStateKey === 'empty'
      ? {}
      : cloneData(dnsRecords[recordStateKey]);

    const data = Object.assign({}, base, derivedState);
    if (base.panel && derivedState.panel) {
      data.panel = Object.assign({}, base.panel, derivedState.panel);
    }

    if (recordStateKey === 'recordAdded' && data.panel && data.panel.record) {
      data.records = [data.panel.record].concat(data.records || []);
    }

    data.infoTerms = ['A', 'AAAA', 'CNAME', 'MX', 'TXT', 'TTL'];
    return data;
  }

  function resolvePolicyPageData(pageType, state = 'default', domain = '') {
    const policyKey = pageType === 'block-list' ? 'blockList' : 'allowList';
    const policy = window.OpenFiltrMockData && window.OpenFiltrMockData[policyKey];
    if (!policy) {
      return null;
    }

    const stateMap = {
      default: 'normal',
      empty: 'empty',
      'add-rule': 'addRule',
      'rule-added': 'ruleAdded',
    };

    const policyStateKey = stateMap[state] || 'normal';
    const base = cloneData(state === 'empty' ? policy.empty : policy.normal);
    const derivedState = policyStateKey === 'normal' || policyStateKey === 'empty'
      ? {}
      : cloneData(policy[policyStateKey]);

    const data = Object.assign({}, base, derivedState);
    if (base.panel && derivedState.panel) {
      data.panel = Object.assign({}, base.panel, derivedState.panel);
    }

    if (data.panel && data.panel.fields && domain) {
      data.panel.fields.domain = domain;
    }

    if (data.panel && data.panel.mode === 'add-rule' && domain) {
      data.panel.domain = domain;
    }

    data.infoTerms = ['allow list', 'block list', 'matched rule'];
    data.pageType = pageType;
    return data;
  }

  function getInfoPanelContent(term = '') {
    const panels = window.OpenFiltrMockData && window.OpenFiltrMockData.infoPanels;
    if (!panels) {
      return null;
    }

    return cloneData(panels[term]) || null;
  }

  function getInfoPanelMarkup(content) {
    if (!content) {
      return '';
    }

    return [
      `<p class="of-info-panel__title">${content.title}</p>`,
      `<p class="of-info-panel__copy">${content.copy}</p>`,
      `<p class="of-info-panel__example">${content.example}</p>`,
    ].join('');
  }

  function bindInfoPanelGroup({ panel, triggers = [] } = {}) {
    if (!panel || !triggers.length) {
      return {
        close() {},
      };
    }

    let activeTrigger = null;

    function close() {
      activeTrigger = null;
      panel.innerHTML = '';
      setHidden(panel, true);
      triggers.forEach((trigger) => {
        trigger.setAttribute('aria-expanded', 'false');
      });
    }

    function open(trigger) {
      const term = trigger.getAttribute('data-info-term');
      const content = getInfoPanelContent(term);

      if (!content) {
        close();
        return;
      }

      activeTrigger = trigger;
      panel.innerHTML = getInfoPanelMarkup(content);
      setHidden(panel, false);

      triggers.forEach((item) => {
        item.setAttribute('aria-expanded', item === trigger ? 'true' : 'false');
      });
    }

    triggers.forEach((trigger) => {
      trigger.addEventListener('click', (event) => {
        event.preventDefault();
        event.stopPropagation();

        if (activeTrigger === trigger && !panel.hidden) {
          close();
          return;
        }

        open(trigger);
      });
    });

    document.addEventListener('click', (event) => {
      if (panel.hidden) {
        return;
      }

      const clickedTrigger = triggers.some((trigger) => trigger.contains(event.target));
      if (clickedTrigger || panel.contains(event.target)) {
        return;
      }

      close();
    });

    document.addEventListener('keydown', (event) => {
      if (event.key === 'Escape' && !panel.hidden) {
        close();
      }
    });

    close();

    return {
      close,
      open,
    };
  }

  function getBadgeClassName(tone = 'neutral') {
    const tones = {
      accent: 'of-badge of-badge--accent',
      neutral: 'of-badge of-badge--neutral',
      success: 'of-badge of-badge--success',
      warning: 'of-badge of-badge--warning',
      danger: 'of-badge of-badge--danger',
    };

    return tones[tone] || tones.neutral;
  }

  window.OpenFiltrPrototype = Object.freeze({
    reviewStates: REVIEW_STATES.slice(),
    getQueryParam,
    getReviewState,
    toggleClass,
    setHidden,
    setReviewStateClass,
    bootPage,
    validateSetupFields,
    resolveLoginAttempt,
    resolveDashboardData,
    resolveActivityData,
    resolveDnsRecordsData,
    resolvePolicyPageData,
    getInfoPanelContent,
    getInfoPanelMarkup,
    bindInfoPanelGroup,
    getBadgeClassName,
  });
}());
