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

  const commonQuickActions = ['Import config', 'View activity', 'Manage rules'];

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
      recentActivity: [
        {
          domain: 'ads.example.net',
          action: 'Blocked',
          actionTone: 'accent',
          client: 'Kitchen tablet',
          category: 'Advertising',
          rule: 'Default block policy',
          time: '1 min ago',
        },
        {
          domain: 'cdn.analytics.example',
          action: 'Blocked',
          actionTone: 'accent',
          client: 'Work laptop',
          category: 'Analytics',
          rule: 'Telemetry block list',
          time: '4 mins ago',
        },
        {
          domain: 'video.example.org',
          action: 'Allowed',
          actionTone: 'neutral',
          client: 'Office Wi-Fi',
          category: 'Streaming',
          rule: 'Allowed by policy',
          time: '7 mins ago',
        },
        {
          domain: 'tracking.example.io',
          action: 'Blocked',
          actionTone: 'accent',
          client: 'Guest phone',
          category: 'Tracking',
          rule: 'Privacy rule set',
          time: '12 mins ago',
        },
      ],
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
      quickActions: commonQuickActions,
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
      quickActions: commonQuickActions,
    },
  };

  window.OpenFiltrMockData = deepFreeze({
    dashboard,
  });
}());
