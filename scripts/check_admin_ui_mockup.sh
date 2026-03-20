#!/usr/bin/env bash

set -u

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PROTOTYPE_DIR="$ROOT_DIR/docs/mockups/admin-ui-v1"
INDEX_HTML="$PROTOTYPE_DIR/index.html"
DASHBOARD_HTML="$PROTOTYPE_DIR/dashboard.html"
SETUP_HTML="$PROTOTYPE_DIR/setup.html"
LOGIN_HTML="$PROTOTYPE_DIR/login.html"
ACTIVITY_HTML="$PROTOTYPE_DIR/activity.html"
DNS_RECORDS_HTML="$PROTOTYPE_DIR/dns-records.html"
ALLOW_LIST_HTML="$PROTOTYPE_DIR/allow-list.html"
BLOCK_LIST_HTML="$PROTOTYPE_DIR/block-list.html"
LOGO="$PROTOTYPE_DIR/assets/openfiltr-logo.svg"
THEME_JS="$PROTOTYPE_DIR/theme.js"
STYLES="$PROTOTYPE_DIR/styles.css"
MOCK_DATA="$PROTOTYPE_DIR/mock-data.js"
PROTOTYPE_JS="$PROTOTYPE_DIR/prototype.js"

fail=0

check_file() {
  local path="$1"
  local label="$2"
  if [[ ! -f "$path" ]]; then
    printf 'Missing %s: %s\n' "$label" "$path" >&2
    fail=1
  fi
}

check_file "$SETUP_HTML" "setup.html"
check_file "$LOGIN_HTML" "login.html"
check_file "$DASHBOARD_HTML" "dashboard.html"
check_file "$INDEX_HTML" "index.html"
check_file "$ACTIVITY_HTML" "activity.html"
check_file "$DNS_RECORDS_HTML" "dns-records.html"
check_file "$ALLOW_LIST_HTML" "allow-list.html"
check_file "$BLOCK_LIST_HTML" "block-list.html"
check_file "$LOGO" "logo asset"
check_file "$THEME_JS" "theme.js"
check_file "$STYLES" "styles.css"
check_file "$MOCK_DATA" "mock-data.js"
check_file "$PROTOTYPE_JS" "prototype.js"

check_unapproved_external_https() {
  local path="$1"

  if [[ ! -f "$path" ]]; then
    return
  fi

  if grep -n 'https://' "$path" | grep -vF 'https://cdn.tailwindcss.com' >/dev/null 2>&1; then
    printf 'Unapproved external HTTPS dependency found in %s\n' "$(basename "$path")" >&2
    grep -n 'https://' "$path" | grep -vF 'https://cdn.tailwindcss.com' >&2
    fail=1
  fi
}

if ! command -v node >/dev/null 2>&1; then
  printf 'Missing required tool: node\n' >&2
  fail=1
fi

check_page() {
  local path="$1"
  local label="$2"
  shift 2

  if [[ ! -f "$path" ]]; then
    return
  fi

  for needle in "$@"; do
    if ! grep -qF "$needle" "$path"; then
      printf 'Missing %s marker in %s: %s\n' "$label" "$(basename "$path")" "$needle" >&2
      fail=1
    fi
  done
}

check_file_contains() {
  local path="$1"
  local label="$2"
  shift 2

  if [[ ! -f "$path" ]]; then
    return
  fi

  for needle in "$@"; do
    if ! grep -qF "$needle" "$path"; then
      printf 'Missing %s marker in %s: %s\n' "$label" "$(basename "$path")" "$needle" >&2
      fail=1
    fi
  done
}

check_file_contains "$THEME_JS" "theme" \
  'window.tailwind = window.tailwind || {};' \
  'window.tailwind.config = {' \
  'theme: {' \
  'colors: {' \
  'canvasSoft' \
  'accentStrong' \
  'fontFamily: {' \
  'spacing: {' \
  'boxShadow: {' \
  'borderRadius: {' \
  'maxWidth: {' \
  'letterSpacing: {' \
  'label'

check_file_contains "$STYLES" "styles" \
  '.of-shell' \
  '.of-topbar' \
  '.of-panel' \
  '.of-section-heading' \
  '.of-button' \
  '.of-button--primary' \
  '.of-button--secondary' \
  '.of-button--ghost' \
  '.of-input' \
  '.of-badge' \
  '.of-list-row' \
  '.of-stepper' \
  '.of-step' \
  '.of-kpi' \
  '.of-app-shell' \
  '.of-side-nav' \
  '.of-toolbar' \
  '.of-table' \
  '.of-info-trigger' \
  '.of-info-panel'

check_file_contains "$MOCK_DATA" "mock data" \
  'dashboard' \
  'activity' \
  'dnsRecords' \
  'allowList' \
  'blockList' \
  'infoPanels'

check_file_contains "$PROTOTYPE_JS" "prototype" \
  'resolveActivityData' \
  'resolveDnsRecordsData' \
  'resolvePolicyPageData' \
  'getInfoPanelContent' \
  'getQueryParam'

check_page "$SETUP_HTML" "setup" \
  'assets/openfiltr-logo.svg' \
  'theme.js' \
  'https://cdn.tailwindcss.com' \
  'styles.css' \
  'prototype.js' \
  'class="of-shell' \
  'class="of-panel' \
  'class="of-input' \
  'class="of-button of-button--primary' \
  'data-step="admin-account"' \
  'data-step="setup-complete"' \
  'data-review-state="validation-error"'

check_page "$LOGIN_HTML" "login" \
  'assets/openfiltr-logo.svg' \
  'theme.js' \
  'https://cdn.tailwindcss.com' \
  'styles.css' \
  'prototype.js' \
  'class="of-shell' \
  'class="of-panel' \
  'class="of-input' \
  'class="of-button of-button--primary' \
  'data-review-state="error"' \
  'data-success-target="dashboard.html"'

check_page "$INDEX_HTML" "launcher" \
  'theme.js' \
  'https://cdn.tailwindcss.com' \
  'class="of-shell' \
  'class="of-panel' \
  'setup.html' \
  'login.html' \
  'dashboard.html' \
  'activity.html' \
  'dns-records.html' \
  'allow-list.html' \
  'block-list.html' \
  'setup.html?state=validation-error' \
  'login.html?state=error' \
  'dashboard.html?state=low-data' \
  'activity.html?filter=blocked' \
  'dns-records.html?state=add-record' \
  'dns-records.html?state=import-preview' \
  'allow-list.html?state=add-rule' \
  'block-list.html?state=add-rule'

check_page "$DASHBOARD_HTML" "dashboard" \
  'assets/openfiltr-logo.svg' \
  'theme.js' \
  'https://cdn.tailwindcss.com' \
  'styles.css' \
  'prototype.js' \
  'class="of-shell' \
  'class="of-topbar' \
  'class="of-panel' \
  'class="of-list-row' \
  'class="of-badge' \
  'data-page="dashboard"' \
  'Service health' \
  'Total requests' \
  'Blocked requests' \
  'Allowed requests' \
  'Block rate' \
  'Recent activity' \
  'Top blocked domains' \
  'Add DNS record' \
  'Allow domain' \
  'Block domain' \
  'Review blocked traffic' \
  'data-ui="metric-tile"'

check_page "$ACTIVITY_HTML" "activity" \
  'theme.js' \
  'styles.css' \
  'prototype.js' \
  'data-page="activity"' \
  'Recent activity' \
  'matched rule' \
  'activity.html?filter=blocked'

check_page "$DNS_RECORDS_HTML" "dns records" \
  'theme.js' \
  'styles.css' \
  'prototype.js' \
  'data-page="dns-records"' \
  'DNS records' \
  'Add record' \
  'A' \
  'AAAA' \
  'CNAME' \
  'TTL'

check_page "$ALLOW_LIST_HTML" "allow list" \
  'theme.js' \
  'styles.css' \
  'prototype.js' \
  'data-page="allow-list"' \
  'Allow list' \
  'Add allow rule'

check_page "$BLOCK_LIST_HTML" "block list" \
  'theme.js' \
  'styles.css' \
  'prototype.js' \
  'data-page="block-list"' \
  'Block list' \
  'Add block rule'

check_unapproved_external_https "$INDEX_HTML"
check_unapproved_external_https "$SETUP_HTML"
check_unapproved_external_https "$LOGIN_HTML"
check_unapproved_external_https "$DASHBOARD_HTML"
check_unapproved_external_https "$ACTIVITY_HTML"
check_unapproved_external_https "$DNS_RECORDS_HTML"
check_unapproved_external_https "$ALLOW_LIST_HTML"
check_unapproved_external_https "$BLOCK_LIST_HTML"
check_unapproved_external_https "$STYLES"
check_unapproved_external_https "$MOCK_DATA"
check_unapproved_external_https "$PROTOTYPE_JS"
check_unapproved_external_https "$THEME_JS"

if command -v node >/dev/null 2>&1; then
  (cd "$ROOT_DIR" && node <<'EOF'
  const fs = require('fs');
  const vm = require('vm');

  global.window = { location: { search: '' } };
  global.document = { addEventListener() {} };
  require('./docs/mockups/admin-ui-v1/mock-data.js');
  require('./docs/mockups/admin-ui-v1/prototype.js');
  const themeSource = fs.readFileSync('./docs/mockups/admin-ui-v1/theme.js', 'utf8');

  vm.runInNewContext(themeSource, { window: global.window });

  const prototype = global.window.OpenFiltrPrototype;
  const theme = global.window.tailwind && global.window.tailwind.config;

  function assert(condition, message) {
    if (!condition) {
      throw new Error(message);
    }
  }

  const invalidSetup = prototype.validateSetupFields({
    username: '',
    password: 'short1',
    confirmPassword: 'short2',
  });

  assert(invalidSetup.isValid === false, 'expected invalid setup fixture to fail');
  assert(invalidSetup.errors.username === 'Enter a username.', 'expected username validation');
  assert(invalidSetup.errors.password === 'Password must be at least 8 characters.', 'expected password length validation');
  assert(invalidSetup.errors.confirmPassword === 'Passwords do not match.', 'expected password mismatch validation');

  const validSetup = prototype.validateSetupFields({
    username: 'admin',
    password: 'password1',
    confirmPassword: 'password1',
  });

  assert(validSetup.isValid === true, 'expected valid setup fixture to pass');

  const forcedLogin = prototype.resolveLoginAttempt({
    forceError: true,
    username: 'admin',
    password: 'password1',
  });

  assert(forcedLogin.status === 'error', 'expected forced login attempt to fail');

  const successfulLogin = prototype.resolveLoginAttempt({
    forceError: false,
    username: 'admin',
    password: 'password1',
  });

  assert(successfulLogin.status === 'success', 'expected valid login attempt to succeed');

  const normalDashboard = prototype.resolveDashboardData('default');
  assert(normalDashboard.health.service === 'Running', 'expected normal dashboard health');
  assert(normalDashboard.recentActivity.length > 0, 'expected normal dashboard activity');

  const lowDataDashboard = prototype.resolveDashboardData('low-data');
  assert(lowDataDashboard.health.service === 'Running', 'expected low-data dashboard health');
  assert(lowDataDashboard.recentActivity.length === 0, 'expected low-data recent activity to be empty');
  assert(lowDataDashboard.topBlockedDomains.length === 0, 'expected low-data blocked domains to be empty');
  assert(lowDataDashboard.emptyState.recentActivity.length > 0, 'expected low-data empty-state copy');

  const activityData = prototype.resolveActivityData('default', '');
  assert(activityData.items.length > 0, 'expected default activity feed');

  const blockedActivity = prototype.resolveActivityData('default', 'blocked');
  assert(blockedActivity.items.every((item) => item.action === 'Blocked'), 'expected blocked activity filter');

  const dnsRecordsData = prototype.resolveDnsRecordsData('default');
  assert(dnsRecordsData.records.length > 0, 'expected populated DNS records state');

  const addRecordState = prototype.resolveDnsRecordsData('add-record');
  assert(addRecordState.panel.mode === 'add-record', 'expected add-record state to open add panel');

  const importPreviewState = prototype.resolveDnsRecordsData('import-preview');
  assert(importPreviewState.panel.mode === 'import-preview', 'expected import-preview state to open import panel');

  const allowListData = prototype.resolvePolicyPageData('allow-list', 'default', '');
  assert(allowListData.items.length > 0, 'expected populated allow list');

  const blockListData = prototype.resolvePolicyPageData('block-list', 'add-rule', 'ads.example.net');
  assert(blockListData.panel.mode === 'add-rule', 'expected add-rule state to open policy panel');
  assert(blockListData.panel.domain === 'ads.example.net', 'expected domain prefill for block rule');

  const infoPanel = prototype.getInfoPanelContent('AAAA');
  assert(infoPanel.title === 'AAAA record', 'expected AAAA info panel title');
  assert(infoPanel.copy.length > 0, 'expected AAAA info panel copy');

  assert(theme && theme.theme && theme.theme.extend, 'expected Tailwind theme config');
  assert(theme.theme.extend.colors.of.accent === '#0b63d1', 'expected accent colour');
  assert(theme.theme.extend.fontFamily.sans[0] === 'ui-sans-serif', 'expected sans stack');
  assert(theme.theme.extend.spacing[18] === '4.5rem', 'expected spacing token');
  assert(theme.theme.extend.boxShadow.panel.includes('rgba(15, 23, 42, 0.06)'), 'expected panel shadow');
  assert(theme.theme.extend.borderRadius.control === '0.9rem', 'expected control radius');
EOF
  )
  if [[ $? -ne 0 ]]; then
    printf 'Prototype behaviour check failed.\n' >&2
    fail=1
  fi
fi

if [[ "$fail" -ne 0 ]]; then
  exit 1
fi

printf 'Admin UI mock-up check passed.\n'
