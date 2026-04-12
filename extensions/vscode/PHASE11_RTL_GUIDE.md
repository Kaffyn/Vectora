# Phase 11: RTL Languages Support & i18n Enhancement

## Overview

Phase 11 introduces comprehensive RTL (Right-to-Left) language support and enhanced internationalization (i18n) capabilities to the Vectora VS Code extension. This phase supports 12 languages across both LTR and RTL text directions with full type safety and automatic direction detection.

## Supported Languages

### LTR Languages (8 + existing)
- **en-US**: English (United States)
- **pt-BR**: Portuguese (Brazil) - Existing
- **pt-PT**: Portuguese (Portugal) - New
- **es-ES**: Spanish (Spain) - Existing
- **fr-FR**: French (France) - New
- **de-DE**: German (Germany) - New
- **it-IT**: Italian (Italy) - New
- **ja-JP**: Japanese (Japan) - New
- **zh-CN**: Chinese Simplified (China) - New
- **zh-TW**: Chinese Traditional (Taiwan) - New
- **ko-KR**: Korean (South Korea) - New

### RTL Languages (4)
- **ar-SA**: Arabic (Saudi Arabia)
- **he-IL**: Hebrew (Israel)
- **fa-IR**: Persian (Iran)
- **ur-PK**: Urdu (Pakistan)

## Architecture

### Directory Structure

```
extensions/vscode/src/
├── i18n/
│   ├── rtl/
│   │   ├── rtlDetector.ts      # RTL language detection
│   │   ├── rtlManager.ts       # RTL state management
│   │   └── rtlUtils.ts         # RTL utility functions
│   ├── translations/
│   │   ├── en-US.json
│   │   ├── pt-BR.json
│   │   ├── pt-PT.json          # NEW
│   │   ├── es-ES.json
│   │   ├── fr-FR.json          # NEW
│   │   ├── de-DE.json          # NEW
│   │   ├── it-IT.json          # NEW
│   │   ├── ja-JP.json          # NEW
│   │   ├── zh-CN.json          # NEW
│   │   ├── zh-TW.json          # NEW
│   │   ├── ko-KR.json          # NEW
│   │   ├── ar-SA.json          # NEW (RTL)
│   │   ├── he-IL.json          # NEW (RTL)
│   │   ├── fa-IR.json          # NEW (RTL)
│   │   └── ur-PK.json          # NEW (RTL)
│   ├── types.ts                # Updated with RTL types
│   ├── index.ts                # Updated with RTL integration
│   └── TranslationContext.tsx   # React context provider
├── styles/rtl/
│   ├── rtl-base.css            # Core RTL styles
│   ├── rtl-components.css      # Component-specific RTL
│   └── rtl-themes.css          # Theme-aware RTL
├── hooks/
│   └── useRTL.ts               # React hook for RTL
└── components/i18n/
    ├── LanguageSelector.tsx    # Enhanced language selector
    └── RTLTest.tsx             # RTL testing component
```

## Core Features

### 1. RTL Detection (`rtlDetector.ts`)

```typescript
import { isRTLLanguage, getTextDirection } from "@/i18n/rtl/rtlDetector";

// Check if a language is RTL
const isArabic = isRTLLanguage("ar-SA"); // true

// Get text direction
const dir = getTextDirection("he-IL"); // "rtl"

// Detect from browser locale
const preferred = getPreferredLanguageFromBrowser(); // "ar-SA" or "en-US"
```

**Key Functions:**
- `isRTLLanguage(language)` - Check if language is RTL
- `isLTRLanguage(language)` - Check if language is LTR
- `getTextDirection(language)` - Get "ltr" or "rtl"
- `getRTLLanguages()` - Get all RTL languages
- `detectRTLFromLocale(locale)` - Detect RTL from browser locale
- `getPreferredLanguageFromBrowser()` - Auto-detect user's preferred language

### 2. RTL Manager (`rtlManager.ts`)

```typescript
import { RTLManager, initRTLManager } from "@/i18n/rtl/rtlManager";

// Initialize
const rtl = initRTLManager("ar-SA", {
  enabled: true,
  rtlClass: "rtl",
  ltrClass: "ltr",
  useCSSLogicalProperties: true,
  autoFlipLayout: true,
  flipIcons: true,
});

// Apply direction to page
rtl.applyDirection("he-IL");

// Check if RTL
if (rtl.isRTL()) {
  console.log("Current language is RTL");
}

// Get direction class
const className = rtl.getDirectionClass(); // "rtl" or "ltr"

// Apply RTL to element
rtl.applyRTLStyles(element);

// Icon flipping
if (rtl.shouldFlipIcon("arrow-left")) {
  const flipped = rtl.getFlippedIconName("arrow-left"); // "arrow-right"
}
```

**Key Methods:**
- `applyDirection(language)` - Set RTL/LTR direction
- `isRTL(language?)` - Check if language is RTL
- `getDirectionClass()` - Get "rtl" or "ltr" class
- `applyRTLStyles(element)` - Apply direction to element
- `removeRTLStyles(element)` - Remove direction styles
- `shouldFlipIcon(iconName)` - Check if icon should flip
- `getFlippedIconName(iconName)` - Get flipped icon name

### 3. RTL Utilities (`rtlUtils.ts`)

```typescript
import {
  flipCSSValue,
  flipStyles,
  getTextAlign,
  getFlexDirection,
  containsRTLCharacters,
  detectDirectionFromContent,
  convertToLogicalProperties,
} from "@/i18n/rtl/rtlUtils";

// Flip CSS values
const value = flipCSSValue("text-align", "left"); // "right"

// Flip entire style object
const styles = flipStyles({ marginLeft: "1rem", textAlign: "left" }, true);
// { marginRight: "1rem", textAlign: "right" }

// Get directional text alignment
const align = getTextAlign(true, "left"); // "right" (RTL)

// Get flex direction
const flex = getFlexDirection(true, "row"); // "row-reverse" (RTL)

// Detect RTL characters
const hasRTL = containsRTLCharacters("مرحبا بك"); // true

// Detect content direction
const direction = detectDirectionFromContent("Hello مرحبا"); // "mixed"

// Convert to CSS logical properties
const logical = convertToLogicalProperties({ marginLeft: "1rem" });
// { marginInlineStart: "1rem" }
```

**Key Utilities:**
- `flipCSSValue(property, value)` - Flip CSS values
- `flipStyles(styles, isRTL)` - Flip entire style objects
- `getDirectionalMargin()` - Direction-aware margin
- `getDirectionalPadding()` - Direction-aware padding
- `getTextAlign(isRTL, align)` - Direction-aware alignment
- `getFlexDirection(isRTL, base)` - Direction-aware flex
- `containsRTLCharacters(text)` - Check for RTL content
- `detectDirectionFromContent(text)` - Auto-detect direction
- `convertToLogicalProperties(styles)` - CSS logical properties

### 4. React Hook: useRTL

```typescript
import { useRTL, useRTLStyles, useLanguageDirection } from "@/hooks/useRTL";

function MyComponent() {
  const { isRTL, direction, language, shouldFlipIcon, getFlippedIcon } = useRTL();
  const { getTextAlign, getFlexDirection, getPosition } = useRTLStyles();
  const { direction: dir, isRTL: rtl } = useLanguageDirection();

  return (
    <div
      style={{
        flexDirection: getFlexDirection("row"),
        textAlign: getTextAlign("left"),
      }}
      dir={direction}
    >
      {shouldFlipIcon("arrow-left") && (
        <Icon name={getFlippedIcon("arrow-left")} />
      )}
    </div>
  );
}
```

**Hooks:**
- `useRTL()` - Full RTL state and utilities
- `useRTLStyles()` - Direction-aware CSS utilities
- `useLanguageDirection()` - Simple direction getter

### 5. CSS RTL Support

#### Base Styles (`rtl-base.css`)

```css
/* CSS Variables for RTL */
:root {
  --text-direction: ltr;
  --start: left;
  --end: right;
  --flex-direction: row;
}

:root[dir="rtl"] {
  --text-direction: rtl;
  --start: right;
  --end: left;
  --flex-direction: row-reverse;
}

/* Logical Properties */
.text-start { text-align: var(--text-align-start); }
.position-start { inset-inline-start: var(--start); }
.margin-start { margin-inline-start: var(--start); }
.padding-start { padding-inline-start: var(--start); }
```

#### Component Styles (`rtl-components.css`)

Automatically handles RTL for:
- Forms and inputs
- Buttons
- Navigation
- Cards and panels
- Lists
- Dropdowns
- Modals
- Badges
- Alerts
- Breadcrumbs
- Pagination
- Sidebars
- Tables
- And more...

#### Theme Styles (`rtl-themes.css`)

Theme-aware RTL styling for:
- Light/Dark/High Contrast modes
- Grid and Flexbox layouts
- Typography
- Animations
- Shadows
- Overflow handling
- Text truncation
- Transitions
- Accessibility

## Language Metadata

Each language includes rich metadata:

```typescript
interface LanguageMetadata {
  code: SupportedLanguage;
  nativeName: string;        // "العربية" or "English"
  englishName: string;       // "Arabic" or "English"
  region: string;            // "Saudi Arabia"
  direction: "ltr" | "rtl";
  dateFormat: string;        // "DD/MM/YYYY"
  timeFormat: "24h" | "12h";
  numberFormat: {
    decimal: string;         // "." or ","
    thousands: string;       // "," or "."
  };
  alternativeLanguages?: SupportedLanguage[]; // ["zh-CN"] for zh-TW
}
```

**Access Metadata:**

```typescript
const metadata = i18n.getLanguageMetadata("ar-SA");
console.log(metadata?.nativeName); // "العربية"
console.log(metadata?.direction);  // "rtl"
console.log(metadata?.dateFormat); // "DD/MM/YYYY"

// Get all languages
const allLanguages = i18n.getAllLanguagesMetadata();
```

## Usage Examples

### Basic Language Switching

```typescript
import { useContext } from "react";
import { TranslationContext } from "@/i18n/TranslationContext";

function App() {
  const context = useContext(TranslationContext);

  const handleLanguageChange = async (lang: SupportedLanguage) => {
    await context?.i18n.changeLanguage(lang);
  };

  return (
    <select onChange={(e) => handleLanguageChange(e.target.value as SupportedLanguage)}>
      <option value="en-US">English</option>
      <option value="ar-SA">العربية</option>
      <option value="he-IL">עברית</option>
    </select>
  );
}
```

### RTL-Aware Component

```typescript
import { useRTL } from "@/hooks/useRTL";

function Card({ title, content }: Props) {
  const { direction, isRTL } = useRTL();

  return (
    <article
      dir={direction}
      style={{
        paddingInlineStart: "1rem",
        paddingInlineEnd: "2rem",
        borderInlineStart: "4px solid blue",
      }}
    >
      <h2>{title}</h2>
      <p>{content}</p>
    </article>
  );
}
```

### Enhanced Language Selector

```typescript
import { LanguageSelector } from "@/components/i18n/LanguageSelector";

function Settings() {
  const [language, setLanguage] = useState<SupportedLanguage>("en-US");

  return (
    <LanguageSelector
      currentLanguage={language}
      availableLanguages={["en-US", "ar-SA", "he-IL", "fr-FR", "ja-JP"]}
      onLanguageChange={async (lang) => setLanguage(lang)}
      showRegion={true}
      showDirection={true}
      showMetadata={true}
    />
  );
}
```

### RTL Testing

```typescript
import { RTLTest } from "@/components/i18n/RTLTest";

function DeveloperPanel() {
  return <RTLTest language="ar-SA" />;
}
```

## CSS Logical Properties

The implementation uses CSS Logical Properties for automatic RTL/LTR handling:

```css
/* Traditional LTR Properties */
margin-left, margin-right
padding-left, padding-right
border-left, border-right
left, right

/* CSS Logical Properties (Auto-adjust for RTL) */
margin-inline-start, margin-inline-end
padding-inline-start, padding-inline-end
border-inline-start, border-inline-end
inset-inline-start, inset-inline-end
```

## Event System

Listen for RTL/Language changes:

```typescript
window.addEventListener("rtl-change", (event: CustomEvent) => {
  const { language, direction, rtl } = event.detail;
  console.log(`Language changed to ${language} (${direction})`);
});
```

## Translation Files Structure

Each translation file follows this structure:

```json
{
  "common": { /* Common UI strings */ },
  "chat": { /* Chat-specific strings */ },
  "settings": { /* Settings strings */ },
  "accessibility": { /* A11y strings */ },
  "errors": { /* Error messages */ },
  "apiConfig": { /* API configuration strings */ },
  "commands": { /* Command descriptions */ }
}
```

## Type Safety

Full TypeScript support with proper type definitions:

```typescript
type SupportedLanguage =
  | "en-US" | "pt-BR" | "pt-PT" | "es-ES"
  | "fr-FR" | "de-DE" | "it-IT"
  | "ja-JP" | "zh-CN" | "zh-TW" | "ko-KR"
  | "ar-SA" | "he-IL" | "fa-IR" | "ur-PK";

interface LanguageMetadata {
  code: SupportedLanguage;
  nativeName: string;
  englishName: string;
  region: string;
  direction: "ltr" | "rtl";
  dateFormat: string;
  timeFormat: "24h" | "12h";
  numberFormat: { decimal: string; thousands: string };
}

interface RTLAPI {
  direction: "ltr" | "rtl";
  isRTL(language?: SupportedLanguage): boolean;
  getDirectionClass(language?: SupportedLanguage): string;
  applyRTLStyles(element: HTMLElement): void;
  removeRTLStyles(element: HTMLElement): void;
}
```

## Best Practices

1. **Always use CSS Logical Properties** where possible
2. **Test with both LTR and RTL** languages
3. **Use useRTL hook** in components for direction awareness
4. **Apply direction early** in the component tree
5. **Test multilingual content** visually
6. **Use direction-aware values** in inline styles
7. **Provide metadata** for language selection UI
8. **Handle icon flipping** automatically
9. **Consider number/date formats** in displays
10. **Test in actual browsers** on mobile devices

## Testing Checklist

- [ ] Language switching works for all 12 languages
- [ ] RTL languages display correctly (Arabic, Hebrew, Persian, Urdu)
- [ ] Icons flip correctly in RTL mode
- [ ] Margins and padding flip in RTL
- [ ] Text alignment adjusts for RTL
- [ ] Flex direction reverses in RTL
- [ ] Document direction attribute updates
- [ ] Document language attribute updates
- [ ] CSS classes apply correctly (.rtl, .ltr)
- [ ] CSS variables update correctly
- [ ] RTL event fires on language change
- [ ] useRTL hook updates on language change
- [ ] Metadata displays correct for each language
- [ ] Mixed RTL/LTR content renders correctly
- [ ] Forms work in RTL mode
- [ ] Navigation works in RTL mode
- [ ] Modals/Overlays position correctly in RTL
- [ ] Tooltips position correctly in RTL
- [ ] Accessibility maintained in RTL
- [ ] Screen readers work in RTL mode

## Performance Considerations

- RTL detection runs once at initialization
- CSS variables eliminate runtime overhead
- No layout thrashing from direction changes
- CSS Logical Properties are native - no polyfills needed
- Icon flipping uses simple string replacement
- Metadata is static and can be cached

## Browser Support

- Modern browsers with CSS Logical Properties support
- Firefox 41+
- Chrome 69+
- Safari 15.4+
- Edge 79+

For older browsers, use the non-logical CSS fallbacks included in rtl-base.css.

## Migration Guide

### Updating Existing Components

```typescript
// Before
<div style={{ marginLeft: "1rem", textAlign: "left" }}>

// After
<div style={{
  marginInlineStart: "1rem",
  textAlign: "var(--text-align-start)",
  direction: useRTL().direction,
}}>
```

### Updating CSS

```css
/* Before */
.container {
  margin-left: 1rem;
  padding-right: 2rem;
}

/* After */
.container {
  margin-inline-start: 1rem;
  padding-inline-end: 2rem;
}
```

## Resources

- [MDN: CSS Logical Properties](https://developer.mozilla.org/en-US/docs/Web/CSS/CSS_Logical_Properties)
- [W3C: Writing Modes](https://www.w3.org/TR/css-writing-modes-3/)
- [Google: Building RTL-Aware Web Apps](https://developers.google.com/web/fundamentals/design-and-ux/building-rtl-aware-web-apps)

## Troubleshooting

### Language not changing
- Check if RTL manager is initialized
- Verify language is in supported list
- Check browser console for errors

### Icons not flipping
- Ensure icon name is in flip list
- Check `flipIcons` is enabled in RTL config
- Verify shouldFlipIcon returns true

### RTL styles not applying
- Check CSS classes are imported
- Verify direction attribute set on element
- Check CSS logical properties browser support

## Future Enhancements

- [ ] Additional languages
- [ ] Regional variants
- [ ] Custom RTL configurations per component
- [ ] RTL animation support
- [ ] RTL layout debugging tools
- [ ] Automatic translation services integration
- [ ] Right-to-left number formatting
- [ ] Bidirectional text support
- [ ] Language-specific keyboard layouts
- [ ] Voice direction detection
