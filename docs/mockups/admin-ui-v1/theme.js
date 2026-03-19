window.tailwind = window.tailwind || {};

window.tailwind.config = {
  theme: {
    extend: {
      colors: {
        of: {
          canvas: '#f4f8fd',
          canvasSoft: '#eaf2fb',
          panel: '#ffffff',
          panelMuted: '#f7fbff',
          line: '#d5dfeb',
          lineStrong: '#b5c5d8',
          ink: '#0f172a',
          muted: '#5b6b82',
          accent: '#0b63d1',
          accentStrong: '#084eaa',
          accentSoft: '#e6f0ff',
          success: '#1f7a4a',
          successSoft: '#e7f6ef',
          warning: '#a96b16',
          warningSoft: '#fff3df',
          danger: '#bf3f31',
          dangerSoft: '#fff0ec',
        },
      },
      fontFamily: {
        sans: [
          'ui-sans-serif',
          'system-ui',
          '-apple-system',
          'BlinkMacSystemFont',
          'Segoe UI',
          'sans-serif',
        ],
        mono: [
          'ui-monospace',
          'SFMono-Regular',
          'SF Mono',
          'Menlo',
          'Monaco',
          'Consolas',
          'Liberation Mono',
          'monospace',
        ],
      },
      spacing: {
        18: '4.5rem',
        22: '5.5rem',
        30: '7.5rem',
      },
      boxShadow: {
        panel: '0 1px 2px rgba(15, 23, 42, 0.05), 0 16px 40px rgba(15, 23, 42, 0.06)',
        topbar: '0 1px 0 rgba(15, 23, 42, 0.05), 0 10px 28px rgba(15, 23, 42, 0.05)',
      },
      borderRadius: {
        panel: '1.25rem',
        control: '0.9rem',
      },
      maxWidth: {
        shell: '78rem',
        auth: '72rem',
      },
      letterSpacing: {
        label: '0.14em',
      },
    },
  },
};
