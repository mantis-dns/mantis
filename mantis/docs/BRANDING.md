# Mantis -- Branding Guide

## 1. Name and Identity

### 1.1 Project Name

- **Name**: Mantis
- **Pronunciation**: MAN-tis
- **Etymology**: The praying mantis -- a patient, precise predator that sits motionless until the moment of capture. Mantis does the same with unwanted DNS traffic: silent, watchful, and lethal to ads and trackers.
- **In code**: `mantis`
- **In prose**: Always "Mantis" with capital M. Never "MANTIS" or "mantis" in marketing text.

### 1.2 Tagline

- **Primary**: Silent predator for your network's DNS
- **Technical**: Go-native DNS sinkhole with DHCP, DoH/DoT, and embedded admin UI
- **Marketing**: Block ads and trackers at the network level. One binary. Zero dependencies.

### 1.3 Elevator Pitch

Mantis is a network-level ad blocker written in Go that replaces Pi-hole with a single, self-contained binary. It operates as a DNS sinkhole -- intercepting and blocking requests to advertising and tracking domains before they reach any device on your network. With built-in DHCP, encrypted DNS (DoH/DoT), recursive resolution, and a modern React admin dashboard, Mantis gives you complete control over your network's DNS from a Raspberry Pi to an enterprise deployment.

## 2. Logo

### 2.1 Concept

The logo combines a mantis silhouette with network/DNS imagery. The mantis head shape is stylized into a geometric shield, suggesting protection. The triangular head of a mantis naturally forms a downward-pointing arrow -- referencing DNS resolution and traffic flowing through the filter.

Visual metaphor: The mantis sits at the center of a network, patiently watching all DNS traffic pass through. What passes is clean. What's caught never reaches the network.

### 2.2 Specifications

- **Primary mark**: Geometric mantis head in shield form with "Mantis" wordmark to the right
- **Icon mark**: Mantis head silhouette only, contained in a rounded square (for favicon, app icon)
- **Wordmark**: "Mantis" in Inter SemiBold, tracking +2%
- **Minimum size**: Icon mark: 16px. Full logo: 120px width.
- **Clear space**: Minimum padding equal to the height of the "M" in the wordmark on all sides

### 2.3 AI Generation Prompt

```
Design a minimal, geometric logo for "Mantis" — a network DNS security tool.

Style: Flat, modern, minimal. Single color (emerald green #10B981) on transparent background.
No gradients, no shadows, no 3D effects.

Subject: Stylized praying mantis head viewed from the front, simplified to geometric shapes.
The triangular head should form a shield or downward-pointing arrow shape.
Eyes represented as two small circles or dots.

Composition: Centered, contained within a rounded square boundary (for icon use).
The mantis shape should be recognizable at 32x32px.

Format: SVG vector. Clean paths, minimal anchor points.
Aspect ratio: 1:1 for icon, 3:1 for horizontal logo with wordmark.

Avoid: Realistic insect detail, multiple colors, text baked into the icon,
organic curves (keep geometric), cute/cartoon style.

Mood: Technical, precise, protective, modern. Think Cloudflare or Tailscale brand aesthetic.
```

## 3. Color Palette

### 3.1 Brand Colors

| Role      | Name     | Hex     | RGB              | Usage                                     |
|-----------|----------|---------|------------------|--------------------------------------------|
| Primary   | Emerald  | #10B981 | rgb(16, 185, 129)| Main buttons, active states, primary CTA    |
| Secondary | Teal     | #14B8A6 | rgb(20, 184, 166)| Hover states, secondary buttons, links      |
| Accent    | Amber    | #F59E0B | rgb(245, 158, 11)| Warnings, blocked query highlights, badges  |

**Rationale**: Emerald green evokes the mantis's natural coloring and connotes "protection" and "allowed/safe." Amber provides high-contrast highlights for blocked items and warnings -- a natural predator-prey color dynamic.

### 3.2 Neutrals (Dark Mode -- Default)

| Role           | Hex     | Usage                              |
|----------------|---------|------------------------------------|
| Text Primary   | #F9FAFB | Body text, headings                |
| Text Secondary | #9CA3AF | Captions, placeholders, metadata   |
| Background     | #0F172A | Page background (slate-900)        |
| Surface        | #1E293B | Cards, panels, sidebar (slate-800) |
| Surface Raised | #334155 | Modals, dropdowns (slate-700)      |
| Border         | #334155 | Dividers, input borders            |

### 3.3 Neutrals (Light Mode)

| Role           | Hex     | Usage                              |
|----------------|---------|------------------------------------|
| Text Primary   | #111827 | Body text, headings                |
| Text Secondary | #6B7280 | Captions, placeholders, metadata   |
| Background     | #F8FAFC | Page background                    |
| Surface        | #FFFFFF | Cards, panels, sidebar             |
| Surface Raised | #F1F5F9 | Modals, dropdowns                  |
| Border         | #E2E8F0 | Dividers, input borders            |

### 3.4 Semantic Colors

| Role    | Hex (Dark) | Hex (Light) | Usage                                   |
|---------|------------|-------------|-----------------------------------------|
| Success | #34D399    | #059669     | Allowed queries, healthy status, enabled |
| Error   | #F87171    | #DC2626     | Blocked queries, errors, critical alerts |
| Warning | #FBBF24    | #D97706     | Cautions, degraded status, expiring      |
| Info    | #60A5FA    | #2563EB     | Informational, cached queries, tips      |

### 3.5 Chart Colors

For query log and statistics charts, use a palette that distinguishes allowed vs. blocked at a glance:

| Data Series       | Hex     | Purpose                     |
|-------------------|---------|-----------------------------|
| Allowed queries   | #10B981 | Primary green (emerald-500) |
| Blocked queries   | #F87171 | Red (red-400)               |
| Cached queries    | #60A5FA | Blue (blue-400)             |
| Forwarded queries | #A78BFA | Purple (violet-400)         |
| DHCP events       | #FBBF24 | Amber (amber-400)           |

### 3.6 CSS Variables

```css
:root {
  /* Brand */
  --color-primary: #10B981;
  --color-secondary: #14B8A6;
  --color-accent: #F59E0B;

  /* Semantic */
  --color-success: #34D399;
  --color-error: #F87171;
  --color-warning: #FBBF24;
  --color-info: #60A5FA;
}

/* Dark mode (default) */
:root {
  --color-bg: #0F172A;
  --color-surface: #1E293B;
  --color-surface-raised: #334155;
  --color-text: #F9FAFB;
  --color-text-secondary: #9CA3AF;
  --color-border: #334155;
}

/* Light mode */
.light {
  --color-bg: #F8FAFC;
  --color-surface: #FFFFFF;
  --color-surface-raised: #F1F5F9;
  --color-text: #111827;
  --color-text-secondary: #6B7280;
  --color-border: #E2E8F0;
  --color-success: #059669;
  --color-error: #DC2626;
  --color-warning: #D97706;
  --color-info: #2563EB;
}
```

### 3.7 Tailwind Config Extension

```javascript
colors: {
  mantis: {
    50:  '#ECFDF5',
    100: '#D1FAE5',
    200: '#A7F3D0',
    300: '#6EE7B7',
    400: '#34D399',
    500: '#10B981', // primary
    600: '#059669',
    700: '#047857',
    800: '#065F46',
    900: '#064E3B',
    950: '#022C22',
  }
}
```

## 4. Typography

### 4.1 Font Stack

| Role     | Font         | Weights    | Fallback                                                        |
|----------|-------------|------------|------------------------------------------------------------------|
| Headings | Inter       | 600, 700   | -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif        |
| Body     | Inter       | 400, 500   | -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif        |
| Code     | JetBrains Mono | 400     | "Fira Code", "Cascadia Code", "Source Code Pro", monospace       |

**Rationale**: Inter is designed for computer screens, open-source, variable weight support. Excellent legibility at small sizes -- critical for dense data tables in the admin UI. JetBrains Mono for code and domain names provides clear distinction between similar characters (0/O, 1/l/I).

### 4.2 Type Scale

| Element     | Size     | Weight | Line Height | Letter Spacing |
|-------------|----------|--------|-------------|----------------|
| H1          | 2rem     | 700    | 1.2         | -0.025em       |
| H2          | 1.5rem   | 600    | 1.3         | -0.02em        |
| H3          | 1.25rem  | 600    | 1.4         | -0.01em        |
| Body        | 0.875rem | 400    | 1.6         | 0              |
| Body Large  | 1rem     | 400    | 1.6         | 0              |
| Small       | 0.75rem  | 400    | 1.5         | 0.01em         |
| Code        | 0.8125rem| 400    | 1.5         | 0              |
| Stat Number | 2.5rem   | 700    | 1.1         | -0.03em        |

**Note:** Base body size is 14px (0.875rem), not 16px. Admin dashboards are data-dense -- smaller base allows more information density while maintaining readability with Inter's excellent small-size rendering.

## 5. Voice and Tone

### 5.1 Personality

- **Technical**: Speaks the language of network admins. Uses correct DNS terminology. Never dumbs down.
  - "Mantis blocked 142,367 queries in the last 24 hours across 47 clients."
- **Direct**: No fluff, no marketing speak. States what things do.
  - "Add upstream DNS servers. Mantis queries them in parallel and uses the fastest response."
- **Confident**: Assertive about what it does well. Honest about limitations.
  - "Mantis handles 100K queries/sec on a 4-core machine."
- **Calm**: Network infrastructure should feel reliable, not exciting. No alarm language for routine events.
  - "Upstream 1.1.1.1 is not responding. Queries are being routed to 8.8.8.8."

### 5.2 Writing Rules

- **Headlines**: Short, descriptive, no punctuation. "Query Log" not "Your Query Log!"
- **Documentation**: Technical but accessible. Define acronyms on first use. Use examples.
- **Error messages**: State what happened, why, and what to do. "Blocklist download failed: connection timeout after 30s. Check the URL or try again."
- **Empty states**: Helpful, not sad. "No queries yet. Configure your devices to use Mantis as their DNS server."
- **Numbers**: Use locale-aware formatting. "142,367 queries" not "142367 queries."
- **No emojis**: Professional, technical tool. No emojis in UI, docs, or README.

### 5.3 Vocabulary

| Prefer             | Avoid                          |
|--------------------|--------------------------------|
| Block              | Filter, deny, reject           |
| Allow              | Whitelist, permit, pass        |
| Blocklist          | Blacklist                      |
| Allowlist          | Whitelist                      |
| Upstream           | Forwarder, external DNS        |
| Client             | Device, host (unless specific) |
| Query              | Request, lookup                |
| Gravity            | Compiled blocklist, database   |
| Rebuild Gravity    | Update blocklists, refresh     |

## 6. Visual Language

### 6.1 Border Radius

| Element          | Radius |
|------------------|--------|
| Buttons          | 6px    |
| Cards/Panels     | 8px    |
| Inputs           | 6px    |
| Modals           | 12px   |
| Badges/Tags      | 4px    |
| Avatars/Icons    | Full   |
| Toggle switches  | Full   |

### 6.2 Shadows

Minimal shadows. Use border + surface color differentiation over shadows.

| Level   | Value                                          | Usage               |
|---------|------------------------------------------------|----------------------|
| None    | none                                            | Most elements       |
| Subtle  | 0 1px 2px rgba(0,0,0,0.1)                      | Dropdowns           |
| Medium  | 0 4px 12px rgba(0,0,0,0.15)                    | Modals              |
| Large   | 0 8px 24px rgba(0,0,0,0.2)                     | Popovers, tooltips  |

### 6.3 Spacing

Base unit: 4px. Scale: 4, 8, 12, 16, 20, 24, 32, 40, 48, 64, 80.

| Context                | Spacing      |
|------------------------|-------------|
| Inline elements        | 4-8px       |
| Form field gap         | 12-16px     |
| Card internal padding  | 16-20px     |
| Section gap            | 24-32px     |
| Page margin            | 24-32px     |
| Sidebar width          | 240px       |
| Sidebar collapsed      | 64px        |

### 6.4 Icons

- **Library**: Lucide Icons (lucide-react)
- **Style**: Outline, 1.5px stroke width
- **Default size**: 20px in navigation, 16px inline, 24px in empty states
- **Color**: Inherits text color (currentColor)

**Rationale**: Lucide is the maintained fork of Feather Icons. Consistent stroke width, MIT license, tree-shakable React components. Matches Inter's clean aesthetic.

### 6.5 Data Visualization

- **Chart library**: Recharts
- **Grid lines**: Subtle (border color at 50% opacity)
- **Axes**: Text secondary color, small font (0.75rem)
- **Tooltips**: Surface raised background, small padding, no border
- **Animation**: Gentle entry animation (300ms ease-out), no hover animations
- **Responsive**: Charts fill container width, minimum height 200px

## 7. Assets Checklist

| Asset           | Format       | Size              | Status |
|-----------------|-------------|-------------------|--------|
| Logo (full)     | SVG + PNG   | Vector / 1024px   | TBD    |
| Icon mark       | SVG + PNG   | 512px, 192px, 64px| TBD    |
| Favicon         | .ico + .png | 32px, 16px        | TBD    |
| Apple Touch     | PNG         | 180px             | TBD    |
| OG Image        | PNG         | 1200x630          | TBD    |
| GitHub Social   | PNG         | 1280x640          | TBD    |
| README Header   | SVG         | 800x200           | TBD    |
| Docker Hub Logo | PNG         | 512x512           | TBD    |
