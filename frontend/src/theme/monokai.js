import { tags as t } from '@lezer/highlight';
import { createTheme } from '@uiw/codemirror-themes';

export const monokai = createTheme({
  theme: 'dark',
  settings: {
    background: '#272822',
    foreground: '#F8F8F2',
    caret: '#F8F8F0',
    selection: '#49483E',
    selectionMatch: '#49483E',
    gutterBackground: '#272822',
    gutterForeground: '#75715E',
    gutterBorder: 'transparent',
    lineHighlight: '#3E3D32',
  },
  styles: [
    { tag: t.comment, color: '#75715E' },
    { tag: t.variableName, color: '#F8F8F2' },
    { tag: [t.string, t.special(t.brace)], color: '#E6DB74' },
    { tag: t.number, color: '#AE81FF' },
    { tag: t.bool, color: '#AE81FF' },
    { tag: t.null, color: '#AE81FF' },
    { tag: t.keyword, color: '#F92672' },
    { tag: t.operator, color: '#F92672' },
    { tag: t.className, color: '#A6E22E' },
    { tag: t.definition(t.typeName), color: '#A6E22E' },
    { tag: t.typeName, color: '#66D9EF' },
    { tag: t.angleBracket, color: '#F8F8F2' },
    { tag: t.tagName, color: '#F92672' },
    { tag: t.attributeName, color: '#A6E22E' },
    { tag: t.propertyName, color: '#A6E22E' },
    { tag: t.function(t.variableName), color: '#A6E22E' },
    { tag: t.regexp, color: '#E6DB74' },
    { tag: t.url, color: '#66D9EF' },
  ],
});
