(function () {
  const REVIEW_STATES = ['validation-error', 'error', 'low-data'];

  function getReviewState(search = window.location.search) {
    const params = new URLSearchParams(search);
    const state = params.get('state');

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
      return dashboard.lowData;
    }

    return dashboard.normal;
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
    getReviewState,
    toggleClass,
    setHidden,
    setReviewStateClass,
    bootPage,
    validateSetupFields,
    resolveLoginAttempt,
    resolveDashboardData,
    getBadgeClassName,
  });
}());
