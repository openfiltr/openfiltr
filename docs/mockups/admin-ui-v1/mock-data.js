(function () {
  function deepFreeze(value) {
    if (!value || typeof value !== 'object' || Object.isFrozen(value)) {
      return value;
    }

    Object.getOwnPropertyNames(value).forEach((key) => {
      deepFreeze(value[key]);
    });

    return Object.freeze(value);
  }

  const quickActions = [
    {
      label: 'Add DNS record',
      href: 'dns-records.html?state=add-record',
      detail: 'Create or update host records without leaving the control plane.',
      tone: 'accent',
    },
    {
      label: 'Allow domain',
      href: 'allow-list.html?state=add-rule&domain=video.example.org',
      detail: 'Add a fast exception when a blocked service needs to pass.',
      tone: 'neutral',
    },
    {
      label: 'Block domain',
      href: 'block-list.html?state=add-rule&domain=ads.example.net',
      detail: 'Push a domain into enforcement from the current operational context.',
      tone: 'danger',
    },
    {
      label: 'Review blocked traffic',
      href: 'activity.html?filter=blocked',
      detail: 'Inspect recent blocked requests and promote them into policy if needed.',
      tone: 'warning',
    },
  ];

  const activityItems = [
    {
      id: 'evt-1001',
      domain: 'ads.example.net',
      action: 'Blocked',
      actionTone: 'accent',
      client: 'Kitchen tablet',
      category: 'Advertising',
      rule: 'Default block policy',
      time: '1 min ago',
      queryType: 'A',
      resolver: 'resolver-lon1-02',
      note: 'Seen repeatedly from a single managed client.',
      recommend: {
        allowHref: 'allow-list.html?state=add-rule&domain=ads.example.net',
        blockHref: 'block-list.html?state=add-rule&domain=ads.example.net',
      },
    },
    {
      id: 'evt-1002',
      domain: 'cdn.analytics.example',
      action: 'Blocked',
      actionTone: 'accent',
      client: 'Work laptop',
      category: 'Analytics',
      rule: 'Telemetry block list',
      time: '4 mins ago',
      queryType: 'CNAME',
      resolver: 'resolver-lon1-01',
      note: 'Policy is working, but the host is still generating background lookups.',
      recommend: {
        allowHref: 'allow-list.html?state=add-rule&domain=cdn.analytics.example',
        blockHref: 'block-list.html?state=add-rule&domain=cdn.analytics.example',
      },
    },
    {
      id: 'evt-1003',
      domain: 'video.example.org',
      action: 'Allowed',
      actionTone: 'neutral',
      client: 'Office Wi-Fi',
      category: 'Streaming',
      rule: 'Allowed by policy',
      time: '7 mins ago',
      queryType: 'AAAA',
      resolver: 'resolver-lon1-03',
      note: 'Allowed after an explicit exception for a known service.',
      recommend: {
        allowHref: 'allow-list.html?state=add-rule&domain=video.example.org',
        blockHref: 'block-list.html?state=add-rule&domain=video.example.org',
      },
    },
    {
      id: 'evt-1004',
      domain: 'tracking.example.io',
      action: 'Blocked',
      actionTone: 'accent',
      client: 'Guest phone',
      category: 'Tracking',
      rule: 'Privacy rule set',
      time: '12 mins ago',
      queryType: 'HTTPS',
      resolver: 'resolver-lon1-02',
      note: 'Blocked by an explicit tracking rule and safe to review for broader enforcement.',
      recommend: {
        allowHref: 'allow-list.html?state=add-rule&domain=tracking.example.io',
        blockHref: 'block-list.html?state=add-rule&domain=tracking.example.io',
      },
    },
  ];

  const dashboard = {
    reviewStates: {
      lowData: 'low-data',
    },
    normal: {
      hero: {
        title: 'OpenFiltr is running',
        copy: 'HTTP API and DNS are both healthy.',
        stateBadge: 'Normal sample',
      },
      environment: {
        name: 'Production',
        region: 'LON1',
        summary: 'Resolver traffic is steady and policy sync is current.',
        resolverFleet: '3 nodes',
        policyVersion: 'Policy set 2026.03.4',
        lastSync: '2 minutes ago',
      },
      health: {
        service: 'Running',
        http: 'Healthy',
        dns: 'Healthy',
      },
      stats: {
        totalRequests: 128430,
        blockedRequests: 15432,
        allowedRequests: 112998,
        blockRate: '12.0%',
      },
      recentActivity: activityItems,
      topBlockedDomains: [
        {
          domain: 'analytics.example.com',
          count: 1280,
          detail: 'Telemetry endpoints across managed clients',
          trend: '+14% vs previous hour',
        },
        {
          domain: 'ads.example.net',
          count: 940,
          detail: 'Advertising network requests',
          trend: '+8% vs previous hour',
        },
        {
          domain: 'tracking.example.io',
          count: 612,
          detail: 'Cross-site tracking and pixel traffic',
          trend: '+3% vs previous hour',
        },
        {
          domain: 'metrics.example.dev',
          count: 455,
          detail: 'Developer telemetry and diagnostics',
          trend: '-2% vs previous hour',
        },
      ],
      quickActions,
      emptyState: {
        recentActivity: 'No requests have been logged yet.',
        topBlockedDomains: 'Blocked domains will appear here once traffic is filtered.',
      },
    },
    lowData: {
      hero: {
        title: 'OpenFiltr is running',
        copy: 'The services are healthy, but there is not enough recent traffic to populate the operational panels.',
        stateBadge: 'Low data review state',
      },
      environment: {
        name: 'Production',
        region: 'LON1',
        summary: 'Policy sync is current and the resolver fleet remains healthy.',
        resolverFleet: '3 nodes',
        policyVersion: 'Policy set 2026.03.4',
        lastSync: '2 minutes ago',
      },
      health: {
        service: 'Running',
        http: 'Healthy',
        dns: 'Healthy',
      },
      stats: {
        totalRequests: 18,
        blockedRequests: 0,
        allowedRequests: 18,
        blockRate: '0%',
      },
      recentActivity: [],
      topBlockedDomains: [],
      emptyState: {
        recentActivity: 'No requests have been logged yet.',
        topBlockedDomains: 'Blocked domains will appear here once traffic is filtered.',
      },
      quickActions,
    },
  };

  const activity = {
    normal: {
      title: 'Recent activity',
      copy: 'Inspect the latest resolver decisions, then jump straight into policy work.',
      filters: [
        { label: 'All traffic', value: 'all' },
        { label: 'Blocked only', value: 'blocked' },
        { label: 'Allowed only', value: 'allowed' },
      ],
      items: activityItems,
      emptyState: 'No recent activity has been recorded.',
    },
    lowData: {
      title: 'Recent activity',
      copy: 'The resolver is healthy, but there is not enough recent traffic to show a meaningful feed.',
      filters: [
        { label: 'All traffic', value: 'all' },
        { label: 'Blocked only', value: 'blocked' },
        { label: 'Allowed only', value: 'allowed' },
      ],
      items: [],
      emptyState: 'No recent requests are available for this review state.',
    },
  };

  const dnsRecords = {
    normal: {
      zone: 'home.arpa',
      summary: 'Manage the records this resolver should authoritatively answer for the local deployment.',
      records: [
        { type: 'A', name: 'router', value: '192.0.2.10', ttl: '300', note: 'Gateway appliance' },
        { type: 'AAAA', name: 'router', value: '2001:db8::10', ttl: '300', note: 'IPv6 gateway' },
        { type: 'CNAME', name: 'media', value: 'nas.home.arpa', ttl: '600', note: 'Friendly alias' },
        { type: 'MX', name: '@', value: '10 mail.home.arpa', ttl: '900', note: 'Mail relay' },
        { type: 'TXT', name: '_status', value: '"openfiltr=managed"', ttl: '300', note: 'Operational marker' },
      ],
      panel: {
        mode: 'none',
        title: 'Record details',
        summary: 'Select a record or open the add panel to review the shape of the edit flow.',
      },
    },
    empty: {
      zone: 'lab.internal',
      summary: 'This zone exists, but no records have been added yet.',
      records: [],
      panel: {
        mode: 'none',
        title: 'Empty zone',
        summary: 'Use the add-record panel to create the first host entry.',
      },
    },
    addRecord: {
      panel: {
        mode: 'add-record',
        title: 'Add record',
        summary: 'Start with the common record types and keep the copy short and operational.',
        fields: {
          type: 'A',
          name: 'printer',
          value: '192.0.2.44',
          ttl: '300',
        },
      },
    },
    validationError: {
      panel: {
        mode: 'add-record',
        title: 'Add record',
        summary: 'Fix the highlighted fields to continue.',
        status: 'error',
        fields: {
          type: 'AAAA',
          name: 'printer',
          value: 'invalid-ipv6',
          ttl: '',
        },
        errors: {
          value: 'Enter a valid IPv6 address.',
          ttl: 'Enter a TTL in seconds.',
        },
      },
    },
    recordAdded: {
      panel: {
        mode: 'record-added',
        title: 'Record added',
        summary: 'The new record is visible in the table and ready for review.',
        status: 'success',
        record: { type: 'A', name: 'printer', value: '192.0.2.44', ttl: '300', note: 'Recently added' },
      },
    },
    importPreview: {
      panel: {
        mode: 'import-preview',
        title: 'Import config preview',
        summary: 'Review the parsed zone additions before applying them to the mock-up table.',
        preview: [
          'A api.home.arpa -> 192.0.2.30',
          'AAAA api.home.arpa -> 2001:db8::30',
          'TXT _sync.home.arpa -> "managed=true"',
        ],
      },
    },
  };

  const allowList = {
    normal: {
      title: 'Allow list',
      copy: 'Explicit exceptions for domains that should bypass broader blocking rules.',
      items: [
        {
          domain: 'video.example.org',
          scope: 'All clients',
          reason: 'Staff streaming service',
          lastHit: '7 mins ago',
          matchedRule: 'Manual exception',
        },
        {
          domain: 'cdn.office-suite.example',
          scope: 'Work laptops',
          reason: 'Business productivity suite',
          lastHit: '11 mins ago',
          matchedRule: 'Office allow list',
        },
      ],
      emptyState: 'No allow rules are currently configured.',
      panel: {
        mode: 'none',
        title: 'Allow rule',
        summary: 'Select a rule or open the add panel to review the exception flow.',
      },
    },
    empty: {
      title: 'Allow list',
      copy: 'No explicit allow rules have been added for this review state.',
      items: [],
      emptyState: 'No allow rules are configured in this review state.',
      panel: {
        mode: 'none',
        title: 'Allow rule',
        summary: 'Open the add panel to create the first exception.',
      },
    },
    addRule: {
      panel: {
        mode: 'add-rule',
        title: 'Add allow rule',
        summary: 'Add a concise exception and scope it narrowly where possible.',
        fields: {
          domain: '',
          scope: 'All clients',
          reason: 'Trusted service',
        },
      },
    },
    ruleAdded: {
      panel: {
        mode: 'rule-added',
        title: 'Allow rule added',
        summary: 'The domain now appears in the exception list.',
        status: 'success',
      },
    },
  };

  const blockList = {
    normal: {
      title: 'Block list',
      copy: 'Domains that should always be denied, even when traffic patterns are noisy or intermittent.',
      items: [
        {
          domain: 'ads.example.net',
          scope: 'All clients',
          reason: 'Advertising network',
          lastHit: '1 min ago',
          matchedRule: 'Default block policy',
        },
        {
          domain: 'tracking.example.io',
          scope: 'Guest devices',
          reason: 'Tracking endpoint',
          lastHit: '12 mins ago',
          matchedRule: 'Privacy rule set',
        },
      ],
      emptyState: 'No explicit block rules are currently configured.',
      panel: {
        mode: 'none',
        title: 'Block rule',
        summary: 'Select a rule or open the add panel to review the enforcement flow.',
      },
    },
    empty: {
      title: 'Block list',
      copy: 'No explicit block rules have been added for this review state.',
      items: [],
      emptyState: 'No block rules are configured in this review state.',
      panel: {
        mode: 'none',
        title: 'Block rule',
        summary: 'Open the add panel to create the first enforced domain rule.',
      },
    },
    addRule: {
      panel: {
        mode: 'add-rule',
        title: 'Add block rule',
        summary: 'Add a domain to enforcement with a short reason reviewers can inspect quickly.',
        fields: {
          domain: '',
          scope: 'All clients',
          reason: 'Policy decision',
        },
      },
    },
    ruleAdded: {
      panel: {
        mode: 'rule-added',
        title: 'Block rule added',
        summary: 'The domain now appears in the enforcement list.',
        status: 'success',
      },
    },
  };

  const infoPanels = {
    A: {
      title: 'A record',
      copy: 'Points a hostname to an IPv4 address.',
      example: 'app.example.com -> 203.0.113.10',
    },
    AAAA: {
      title: 'AAAA record',
      copy: 'Points a hostname to an IPv6 address.',
      example: 'app.example.com -> 2001:db8::10',
    },
    CNAME: {
      title: 'CNAME record',
      copy: 'Aliases one hostname to another hostname.',
      example: 'media.example.com -> nas.example.com',
    },
    MX: {
      title: 'MX record',
      copy: 'Declares which mail host should receive mail for a domain.',
      example: 'example.com -> 10 mail.example.com',
    },
    TXT: {
      title: 'TXT record',
      copy: 'Stores free-form text used by verification, routing, or service metadata.',
      example: '_status.example.com -> "managed=true"',
    },
    TTL: {
      title: 'TTL',
      copy: 'Sets how long resolvers may cache the answer before asking again.',
      example: '300 means the answer may be cached for five minutes.',
    },
    'allow list': {
      title: 'Allow list',
      copy: 'Domains here bypass broader blocking rules when they need an explicit exception.',
      example: 'Allowing a business video platform while keeping tracking rules elsewhere.',
    },
    'block list': {
      title: 'Block list',
      copy: 'Domains here are always denied for the selected scope.',
      example: 'Blocking a persistent advertising or tracking endpoint.',
    },
    'matched rule': {
      title: 'Matched rule',
      copy: 'The policy entry that authorised or denied the request.',
      example: 'Default block policy or a manually-added exception.',
    },
  };

  window.OpenFiltrMockData = deepFreeze({
    dashboard,
    activity,
    dnsRecords,
    allowList,
    blockList,
    infoPanels,
  });
}());
