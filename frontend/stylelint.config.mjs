const createGroup = (groupName, properties) => ({
  groupName,
  emptyLineBefore: 'always',
  noEmptyLineBetween: true,
  properties,
});

export default {
  extends: ['stylelint-config-standard-scss'],

  plugins: ['stylelint-order'],

  ignoreFiles: [
    '**/node_modules/**',
    '**/dist/**',
    '**/build/**',
    '**/coverage/**',
    '**/*.min.css',
  ],

  rules: {
    'declaration-empty-line-before': null,
    'selector-class-pattern': null,
    'custom-property-pattern': null,

    'order/properties-order': [
      [
        'all',

        createGroup('Позиционирование', [
          'position',

          'inset',
          'inset-block',
          'inset-block-start',
          'inset-block-end',
          'inset-inline',
          'inset-inline-start',
          'inset-inline-end',

          'top',
          'right',
          'bottom',
          'left',

          'z-index',
        ]),

        createGroup('Блочная модель', [
          'display',

          'float',
          'clear',

          'box-sizing',

          'columns',
          'column-width',
          'column-count',
          'column-fill',
          'column-span',
          'column-rule',
          'column-rule-width',
          'column-rule-style',
          'column-rule-color',

          'flex',
          'flex-grow',
          'flex-shrink',
          'flex-basis',
          'flex-flow',
          'flex-direction',
          'flex-wrap',

          'grid',
          'grid-template',
          'grid-template-areas',
          'grid-template-rows',
          'grid-template-columns',
          'grid-auto-flow',
          'grid-auto-rows',
          'grid-auto-columns',

          'grid-area',
          'grid-row',
          'grid-row-start',
          'grid-row-end',
          'grid-column',
          'grid-column-start',
          'grid-column-end',

          'place-content',
          'align-content',
          'justify-content',

          'place-items',
          'align-items',
          'justify-items',

          'place-self',
          'align-self',
          'justify-self',

          'order',

          'gap',
          'row-gap',
          'column-gap',

          'width',
          'min-width',
          'max-width',

          'height',
          'min-height',
          'max-height',

          'aspect-ratio',

          'margin',
          'margin-block',
          'margin-block-start',
          'margin-block-end',
          'margin-inline',
          'margin-inline-start',
          'margin-inline-end',
          'margin-top',
          'margin-right',
          'margin-bottom',
          'margin-left',

          'padding',
          'padding-block',
          'padding-block-start',
          'padding-block-end',
          'padding-inline',
          'padding-inline-start',
          'padding-inline-end',
          'padding-top',
          'padding-right',
          'padding-bottom',
          'padding-left',

          'overflow',
          'overflow-x',
          'overflow-y',
          'overflow-block',
          'overflow-inline',

          'overscroll-behavior',
          'overscroll-behavior-x',
          'overscroll-behavior-y',

          'scrollbar-gutter',
          'resize',
        ]),

        createGroup('Типографика', [
          'font',
          'font-style',
          'font-variant',
          'font-weight',
          'font-stretch',
          'font-size',
          'line-height',
          'font-family',

          'font-feature-settings',
          'font-kerning',
          'font-optical-sizing',
          'font-synthesis',
          'font-variation-settings',

          'letter-spacing',
          'word-spacing',

          'text-align',
          'text-align-last',
          'text-indent',
          'text-justify',
          'text-transform',

          'text-decoration',
          'text-decoration-line',
          'text-decoration-style',
          'text-decoration-color',
          'text-decoration-thickness',
          'text-underline-offset',

          'text-overflow',
          'text-shadow',
          'text-wrap',

          'white-space',
          'word-break',
          'overflow-wrap',
          'hyphens',
          'tab-size',

          'list-style',
          'list-style-position',
          'list-style-type',
          'list-style-image',

          'direction',
          'writing-mode',

          'color',
        ]),

        createGroup('Оформление', [
          'appearance',
          'accent-color',
          'caret-color',
          'color-scheme',

          'background',
          'background-color',
          'background-image',
          'background-repeat',
          'background-position',
          'background-size',
          'background-attachment',
          'background-origin',
          'background-clip',
          'background-blend-mode',

          'border',
          'border-width',
          'border-style',
          'border-color',

          'border-block',
          'border-block-width',
          'border-block-style',
          'border-block-color',

          'border-inline',
          'border-inline-width',
          'border-inline-style',
          'border-inline-color',

          'border-top',
          'border-top-width',
          'border-top-style',
          'border-top-color',

          'border-right',
          'border-right-width',
          'border-right-style',
          'border-right-color',

          'border-bottom',
          'border-bottom-width',
          'border-bottom-style',
          'border-bottom-color',

          'border-left',
          'border-left-width',
          'border-left-style',
          'border-left-color',

          'border-radius',
          'border-start-start-radius',
          'border-start-end-radius',
          'border-end-start-radius',
          'border-end-end-radius',
          'border-top-left-radius',
          'border-top-right-radius',
          'border-bottom-right-radius',
          'border-bottom-left-radius',

          'outline',
          'outline-width',
          'outline-style',
          'outline-color',
          'outline-offset',

          'box-shadow',
          'opacity',

          'mix-blend-mode',
          'filter',
          'backdrop-filter',

          'clip',
          'clip-path',

          'mask',
          'mask-image',
          'mask-position',
          'mask-size',
          'mask-repeat',

          'object-fit',
          'object-position',
          'image-rendering',

          'fill',
          'fill-opacity',
          'stroke',
          'stroke-width',
          'stroke-linecap',
          'stroke-linejoin',

          'content',
          'quotes',
          'counter-reset',
          'counter-increment',
        ]),

        createGroup('Анимация', [
          'transform',
          'transform-origin',
          'transform-style',

          'translate',
          'rotate',
          'scale',

          'perspective',
          'perspective-origin',
          'backface-visibility',

          'transition',
          'transition-property',
          'transition-duration',
          'transition-timing-function',
          'transition-delay',

          'animation',
          'animation-name',
          'animation-duration',
          'animation-timing-function',
          'animation-delay',
          'animation-iteration-count',
          'animation-direction',
          'animation-fill-mode',
          'animation-play-state',

          'scroll-behavior',
          'scroll-snap-type',
          'scroll-snap-align',
          'scroll-snap-stop',
        ]),

        createGroup('Разное', [
          'cursor',
          'pointer-events',
          'touch-action',
          'user-select',

          'visibility',
          'content-visibility',

          'contain',
          'contain-intrinsic-size',
          'isolation',

          'will-change',
        ]),
      ],

      {
        unspecified: 'bottom',
        emptyLineBeforeUnspecified: 'always',
      },
    ],
  },
};
