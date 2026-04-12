# Phase 11: Supported Languages Reference

## Language Summary

### Total: 12 Languages
- **LTR Languages**: 11 (including English and Spanish)
- **RTL Languages**: 4 (Arabic, Hebrew, Persian, Urdu)

---

## LTR Languages

### English (en-US)
- **Native Name**: English
- **Region**: United States
- **Direction**: LTR
- **Date Format**: MM/DD/YYYY
- **Time Format**: 12h
- **Number Format**: Decimal `.`, Thousands `,`
- **Status**: Existing / Updated
- **Translation Status**: ✓ Complete

### Portuguese (pt-BR)
- **Native Name**: Português
- **Region**: Brazil
- **Direction**: LTR
- **Date Format**: DD/MM/YYYY
- **Time Format**: 24h
- **Number Format**: Decimal `,`, Thousands `.`
- **Status**: Existing / Enhanced
- **Translation Status**: ✓ Complete

### Portuguese (pt-PT)
- **Native Name**: Português
- **Region**: Portugal
- **Direction**: LTR
- **Date Format**: DD/MM/YYYY
- **Time Format**: 24h
- **Number Format**: Decimal `,`, Thousands `.`
- **Status**: New
- **Alternative Languages**: pt-BR
- **Translation Status**: ✓ Complete

### Spanish (es-ES)
- **Native Name**: Español
- **Region**: Spain
- **Direction**: LTR
- **Date Format**: DD/MM/YYYY
- **Time Format**: 24h
- **Number Format**: Decimal `,`, Thousands `.`
- **Status**: Existing / Updated
- **Translation Status**: ✓ Complete

### French (fr-FR)
- **Native Name**: Français
- **Region**: France
- **Direction**: LTR
- **Date Format**: DD/MM/YYYY
- **Time Format**: 24h
- **Number Format**: Decimal `,`, Thousands ` ` (space)
- **Status**: New
- **Translation Status**: ✓ Complete

### German (de-DE)
- **Native Name**: Deutsch
- **Region**: Germany
- **Direction**: LTR
- **Date Format**: DD.MM.YYYY
- **Time Format**: 24h
- **Number Format**: Decimal `,`, Thousands `.`
- **Status**: New
- **Translation Status**: ✓ Complete

### Italian (it-IT)
- **Native Name**: Italiano
- **Region**: Italy
- **Direction**: LTR
- **Date Format**: DD/MM/YYYY
- **Time Format**: 24h
- **Number Format**: Decimal `,`, Thousands `.`
- **Status**: New
- **Translation Status**: ✓ Complete

### Japanese (ja-JP)
- **Native Name**: 日本語
- **Region**: Japan
- **Direction**: LTR
- **Date Format**: YYYY/MM/DD
- **Time Format**: 24h
- **Number Format**: Decimal `.`, Thousands `,`
- **Status**: New
- **Language Family**: CJK (Chinese-Japanese-Korean)
- **Translation Status**: ✓ Complete

### Chinese Simplified (zh-CN)
- **Native Name**: 简体中文
- **Region**: China
- **Direction**: LTR
- **Date Format**: YYYY-MM-DD
- **Time Format**: 24h
- **Number Format**: Decimal `.`, Thousands `,`
- **Status**: New
- **Language Family**: CJK
- **Alternative Languages**: zh-TW
- **Translation Status**: ✓ Complete

### Chinese Traditional (zh-TW)
- **Native Name**: 繁體中文
- **Region**: Taiwan
- **Direction**: LTR
- **Date Format**: YYYY-MM-DD
- **Time Format**: 24h
- **Number Format**: Decimal `.`, Thousands `,`
- **Status**: New
- **Language Family**: CJK
- **Alternative Languages**: zh-CN
- **Translation Status**: ✓ Complete

### Korean (ko-KR)
- **Native Name**: 한국어
- **Region**: South Korea
- **Direction**: LTR
- **Date Format**: YYYY-MM-DD
- **Time Format**: 24h
- **Number Format**: Decimal `.`, Thousands `,`
- **Status**: New
- **Language Family**: CJK
- **Translation Status**: ✓ Complete

---

## RTL Languages

### Arabic (ar-SA)
- **Native Name**: العربية
- **Region**: Saudi Arabia
- **Direction**: RTL
- **Date Format**: DD/MM/YYYY
- **Time Format**: 24h
- **Number Format**: Decimal `,`, Thousands `.`
- **Status**: New
- **Script**: Arabic script
- **Speakers**: ~300+ million
- **Translation Status**: ✓ Complete
- **RTL Features**:
  - Full right-to-left text flow
  - Automatic character shaping
  - Proper number direction handling
  - RTL form input support

### Hebrew (he-IL)
- **Native Name**: עברית
- **Region**: Israel
- **Direction**: RTL
- **Date Format**: DD.MM.YYYY
- **Time Format**: 24h
- **Number Format**: Decimal `.`, Thousands `,`
- **Status**: New
- **Script**: Hebrew script
- **Speakers**: ~9+ million
- **Translation Status**: ✓ Complete
- **RTL Features**:
  - Pure RTL text rendering
  - Diacritical mark support
  - Mixed LTR/RTL content handling
  - Proper punctuation placement

### Persian (fa-IR)
- **Native Name**: فارسی
- **Region**: Iran
- **Direction**: RTL
- **Date Format**: YYYY/MM/DD
- **Time Format**: 24h
- **Number Format**: Decimal `/`, Thousands `,`
- **Status**: New
- **Script**: Persian script (modified Arabic)
- **Speakers**: ~100+ million
- **Translation Status**: ✓ Complete
- **RTL Features**:
  - RTL text with Arabic script variations
  - Proper ligature support
  - Form input RTL handling
  - Numeric display adjustment

### Urdu (ur-PK)
- **Native Name**: اردو
- **Region**: Pakistan
- **Direction**: RTL
- **Date Format**: DD/MM/YYYY
- **Time Format**: 24h
- **Number Format**: Decimal `.`, Thousands `,`
- **Status**: New
- **Script**: Urdu script (Nastaliq or Naskh)
- **Speakers**: ~200+ million
- **Translation Status**: ✓ Complete
- **RTL Features**:
  - RTL text with Urdu-specific characters
  - Proper diacritical marks
  - RTL navigation support
  - Form handling for RTL input

---

## Translation Coverage

### Namespaces Translated
All 12 languages include translations for:

1. **common** (20 strings)
   - UI controls: Cancel, Confirm, OK, Close, etc.
   - Navigation: Back, Next, Previous, etc.
   - General: Search, Settings, About, Help, etc.

2. **chat** (11 strings)
   - Chat interface: Title, Placeholder, Send, etc.
   - Messages: Thinking, User/Assistant labels, etc.
   - Errors: Connection errors, etc.

3. **settings** (12 strings)
   - Settings UI: Language, Theme, Font Size, etc.
   - Options: Dark/Light/High Contrast modes, etc.
   - Accessibility: Keyboard, Screen Reader, etc.

4. **accessibility** (6 strings)
   - A11y features: Keyboard navigation, etc.
   - Screen reader support
   - Accessibility shortcuts

5. **errors** (6 strings)
   - Generic error messages
   - Network, timeout, validation errors
   - Authorization messages

6. **apiConfig** (6 strings)
   - API configuration UI
   - Connection testing
   - Key management

7. **commands** (5 strings)
   - Command descriptions
   - Code operations: Explain, Refactor, etc.
   - Test generation, optimization

**Total Strings Per Language**: 66
**Total Translations**: 792 (12 languages × 66 strings)

---

## Regional Variants

Some languages have regional variants:

### Portuguese
- **pt-BR** (Brazil) - Primary variant with ~215M speakers
- **pt-PT** (Portugal) - Secondary variant with ~10M speakers
- **Alternative mapping**: pt-BR ↔ pt-PT

### Chinese
- **zh-CN** (Simplified) - Mainland China, ~1B speakers
- **zh-TW** (Traditional) - Taiwan, ~23M speakers
- **Alternative mapping**: zh-CN ↔ zh-TW

---

## Language Implementation Details

### Translation File Path
```
extensions/vscode/src/i18n/translations/{code}.json
```

### Loading Strategy
- Languages loaded on-demand when selected
- Cached after first load
- Fallback to English (en-US) for missing translations

### Language Detection
- Browser language preference (navigator.language)
- User configuration
- Fallback to English

---

## Font and Display Recommendations

### LTR Languages
- **Standard fonts**: Arial, Helvetica, Segoe UI, etc.
- **Web safe fonts**: Sans-serif preferred
- **Font weight**: 400-600 recommended

### CJK Languages (ja-JP, zh-CN, zh-TW, ko-KR)
- **Fonts**: Noto Sans CJK, Source Han Sans, etc.
- **Font weight**: 400-500 for body text
- **Line height**: 1.6-1.8 recommended (larger than LTR)
- **Character spacing**: May need adjustment

### RTL Languages (ar-SA, he-IL, fa-IR, ur-PK)
- **Arabic/Persian/Urdu fonts**: Noto Sans Arabic, Droid Arabic, etc.
- **Hebrew fonts**: Noto Sans Hebrew, Noto Sans Mono, etc.
- **Font weight**: 400 recommended
- **Line height**: 1.5-1.6 recommended
- **Letter spacing**: May need adjustment for clarity

---

## Language-Specific Considerations

### Number Formatting
```
English (en-US):     1,234.56
European (de, fr):   1.234,56
Arabic (ar-SA):      ١٬٢٣٤٫٥٦ (Arabic numerals)
Persian (fa-IR):     ۱٬۲۳۴٫۵۶ (Persian numerals)
```

### Date Formatting
```
English (en-US):     12/31/2024
European (pt, es):   31/12/2024
ISO Standard (ja):   2024-12-31
Persian (fa-IR):     1403/10/10 (Solar Hijri)
```

### Time Formatting
```
12-hour (en-US):     3:30 PM
24-hour (most):      15:30
```

---

## Browser Locale Mapping

### Supported Locale Codes
```
en, en-US, en-GB → en-US
pt, pt-BR → pt-BR
pt-PT → pt-PT
es, es-ES → es-ES
fr, fr-FR → fr-FR
de, de-DE → de-DE
it, it-IT → it-IT
ja, ja-JP → ja-JP
zh, zh-CN, zh-Hans → zh-CN
zh-TW, zh-Hant → zh-TW
ko, ko-KR → ko-KR
ar, ar-SA, ar-AE → ar-SA
he, he-IL → he-IL
fa, fa-IR → fa-IR
ur, ur-PK → ur-PK
```

---

## Quality Assurance

### Translation Quality Checks
- ✓ All strings translated (no missing keys)
- ✓ Proper RTL markup for RTL languages
- ✓ Appropriate terminology for technical terms
- ✓ Consistent punctuation and capitalization
- ✓ Proper Unicode characters used

### Visual Quality Checks
- ✓ Text renders correctly in UI
- ✓ No text overflow in UI elements
- ✓ Proper character shaping (Arabic)
- ✓ Correct ligatures where applicable
- ✓ Proper diacritical marks

### Functional Quality Checks
- ✓ Language switching works smoothly
- ✓ RTL layout flips correctly
- ✓ Form inputs work in RTL
- ✓ Navigation works in RTL
- ✓ Modals position correctly

---

## Adding New Languages

To add a new language:

1. **Create translation file**: `src/i18n/translations/{code}.json`
2. **Add to types**: Update `SupportedLanguage` type in `types.ts`
3. **Update language list**: Modify `getAvailableLanguages()` in `index.ts`
4. **Add metadata**: Include in language metadata database
5. **Test thoroughly**: Verify all strings and layouts
6. **Document**: Add to this file and README

---

## Community Contributions

Native speakers are welcome to:
- Review and improve existing translations
- Add new languages
- Report localization issues
- Suggest language-specific improvements

---

## References

- ISO 639-1: Language codes
- ISO 3166-1: Country codes
- Unicode: Character encoding standards
- W3C: Internationalization guidelines
- CLDR: Common Locale Data Repository
