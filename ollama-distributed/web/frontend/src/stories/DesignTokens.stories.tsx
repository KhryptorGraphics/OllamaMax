import { Meta } from '@storybook/react'
import { colors } from '../design-system/tokens/colors'
import { typography } from '../design-system/tokens/typography'
import { spacing } from '../design-system/tokens/spacing'
import { radius } from '../design-system/tokens/radius'
import { shadows } from '../design-system/tokens/shadows'

export default {
  title: 'Design System/Design Tokens',
  parameters: {
    docs: {
      page: () => (
        <div className="p-6">
          <h1 className="text-3xl font-bold mb-6">Design Tokens</h1>
          <p className="text-lg text-muted-foreground mb-8">
            Design tokens are the visual design atoms of the design system — specifically, 
            they are named entities that store visual design attributes.
          </p>

          <div className="space-y-12">
            {/* Colors */}
            <section>
              <h2 className="text-2xl font-semibold mb-6">Colors</h2>
              
              <div className="space-y-8">
                <div>
                  <h3 className="text-lg font-medium mb-4">Primary Colors</h3>
                  <div className="grid grid-cols-2 md:grid-cols-5 lg:grid-cols-10 gap-4">
                    {Object.entries(colors.primary).map(([shade, value]) => (
                      <div key={shade} className="space-y-2">
                        <div 
                          className="w-full h-16 rounded-lg border shadow-sm"
                          style={{ backgroundColor: value }}
                        />
                        <div className="text-sm">
                          <div className="font-mono text-xs">{shade}</div>
                          <div className="font-mono text-xs text-muted-foreground">{value}</div>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>

                <div>
                  <h3 className="text-lg font-medium mb-4">Semantic Colors</h3>
                  <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
                    {['success', 'warning', 'destructive', 'info'].map((colorType) => (
                      <div key={colorType}>
                        <h4 className="font-medium mb-3 capitalize">{colorType}</h4>
                        <div className="space-y-2">
                          {Object.entries(colors[colorType as keyof typeof colors] || {}).map(([shade, value]) => (
                            <div key={shade} className="flex items-center space-x-3">
                              <div 
                                className="w-8 h-8 rounded border"
                                style={{ backgroundColor: value }}
                              />
                              <div className="text-sm">
                                <div className="font-mono">{shade}</div>
                                <div className="font-mono text-xs text-muted-foreground">{value}</div>
                              </div>
                            </div>
                          ))}
                        </div>
                      </div>
                    ))}
                  </div>
                </div>

                <div>
                  <h3 className="text-lg font-medium mb-4">Neutral Colors</h3>
                  <div className="grid grid-cols-2 md:grid-cols-5 lg:grid-cols-10 gap-4">
                    {Object.entries(colors.neutral).map(([shade, value]) => (
                      <div key={shade} className="space-y-2">
                        <div 
                          className="w-full h-16 rounded-lg border"
                          style={{ backgroundColor: value }}
                        />
                        <div className="text-sm">
                          <div className="font-mono text-xs">{shade}</div>
                          <div className="font-mono text-xs text-muted-foreground">{value}</div>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            </section>

            {/* Typography */}
            <section>
              <h2 className="text-2xl font-semibold mb-6">Typography</h2>
              
              <div className="space-y-6">
                <div>
                  <h3 className="text-lg font-medium mb-4">Font Sizes</h3>
                  <div className="space-y-4">
                    {Object.entries(typography.fontSize).map(([size, [fontSize, lineHeight]]) => (
                      <div key={size} className="flex items-center space-x-6 py-2 border-b">
                        <div className="w-16 text-sm font-mono text-muted-foreground">
                          {size}
                        </div>
                        <div className="w-24 text-sm font-mono text-muted-foreground">
                          {fontSize}
                        </div>
                        <div 
                          className="flex-1 font-medium"
                          style={{ 
                            fontSize: fontSize,
                            lineHeight: typeof lineHeight === 'object' ? lineHeight.lineHeight : lineHeight
                          }}
                        >
                          The quick brown fox jumps over the lazy dog
                        </div>
                      </div>
                    ))}
                  </div>
                </div>

                <div>
                  <h3 className="text-lg font-medium mb-4">Font Weights</h3>
                  <div className="space-y-3">
                    {Object.entries(typography.fontWeight).map(([weight, value]) => (
                      <div key={weight} className="flex items-center space-x-6">
                        <div className="w-24 text-sm font-mono text-muted-foreground">
                          {weight}
                        </div>
                        <div className="w-16 text-sm font-mono text-muted-foreground">
                          {value}
                        </div>
                        <div 
                          className="text-lg"
                          style={{ fontWeight: value }}
                        >
                          The quick brown fox jumps over the lazy dog
                        </div>
                      </div>
                    ))}
                  </div>
                </div>

                <div>
                  <h3 className="text-lg font-medium mb-4">Line Heights</h3>
                  <div className="space-y-3">
                    {Object.entries(typography.lineHeight).map(([height, value]) => (
                      <div key={height} className="border rounded-lg p-4">
                        <div className="text-sm font-mono text-muted-foreground mb-2">
                          {height}: {value}
                        </div>
                        <div 
                          className="text-base"
                          style={{ lineHeight: value }}
                        >
                          Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            </section>

            {/* Spacing */}
            <section>
              <h2 className="text-2xl font-semibold mb-6">Spacing</h2>
              <p className="text-muted-foreground mb-6">
                Our spacing system is based on an 8-point grid, providing consistent rhythm and alignment.
              </p>
              
              <div className="space-y-4">
                {Object.entries(spacing).map(([size, value]) => (
                  <div key={size} className="flex items-center space-x-6 py-2">
                    <div className="w-16 text-sm font-mono">
                      {size}
                    </div>
                    <div className="w-20 text-sm font-mono text-muted-foreground">
                      {value}
                    </div>
                    <div className="flex items-center space-x-2">
                      <div 
                        className="bg-primary h-4"
                        style={{ width: value }}
                      />
                      <span className="text-sm text-muted-foreground">
                        {parseInt(value) / 4}rem
                      </span>
                    </div>
                  </div>
                ))}
              </div>
            </section>

            {/* Border Radius */}
            <section>
              <h2 className="text-2xl font-semibold mb-6">Border Radius</h2>
              
              <div className="grid grid-cols-2 md:grid-cols-4 gap-6">
                {Object.entries(radius).map(([size, value]) => (
                  <div key={size} className="space-y-3">
                    <div 
                      className="w-full h-24 bg-muted border-2 border-primary"
                      style={{ borderRadius: value }}
                    />
                    <div className="text-center">
                      <div className="font-mono text-sm">{size}</div>
                      <div className="font-mono text-xs text-muted-foreground">{value}</div>
                    </div>
                  </div>
                ))}
              </div>
            </section>

            {/* Shadows */}
            <section>
              <h2 className="text-2xl font-semibold mb-6">Shadows</h2>
              <p className="text-muted-foreground mb-6">
                Elevation system using shadows to create depth and hierarchy.
              </p>
              
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                {Object.entries(shadows).map(([size, value]) => (
                  <div key={size} className="space-y-3">
                    <div 
                      className="w-full h-24 bg-background border rounded-lg"
                      style={{ boxShadow: value }}
                    />
                    <div className="text-center">
                      <div className="font-mono text-sm">{size}</div>
                      <div className="font-mono text-xs text-muted-foreground break-all">
                        {value}
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </section>

            {/* Usage Guidelines */}
            <section>
              <h2 className="text-2xl font-semibold mb-6">Usage Guidelines</h2>
              
              <div className="space-y-6">
                <div className="p-6 border rounded-lg">
                  <h3 className="text-lg font-medium mb-3">Color Usage</h3>
                  <ul className="space-y-2 text-muted-foreground">
                    <li>• Use semantic colors for status and feedback</li>
                    <li>• Maintain 4.5:1 contrast ratio for normal text</li>
                    <li>• Use neutral colors for UI structure</li>
                    <li>• Primary colors for brand and interactive elements</li>
                  </ul>
                </div>

                <div className="p-6 border rounded-lg">
                  <h3 className="text-lg font-medium mb-3">Typography</h3>
                  <ul className="space-y-2 text-muted-foreground">
                    <li>• Use consistent font weights for hierarchy</li>
                    <li>• Match line heights to content purpose</li>
                    <li>• Scale font sizes for different screen sizes</li>
                    <li>• Maintain readability across all sizes</li>
                  </ul>
                </div>

                <div className="p-6 border rounded-lg">
                  <h3 className="text-lg font-medium mb-3">Spacing</h3>
                  <ul className="space-y-2 text-muted-foreground">
                    <li>• Use 8-point grid for consistent rhythm</li>
                    <li>• Group related elements with closer spacing</li>
                    <li>• Separate unrelated content with larger gaps</li>
                    <li>• Consider touch target sizes for mobile</li>
                  </ul>
                </div>
              </div>
            </section>
          </div>
        </div>
      )
    }
  }
} as Meta

export const ColorPalette = () => <div>See docs panel for color palette</div>
export const TypographySystem = () => <div>See docs panel for typography system</div>
export const SpacingSystem = () => <div>See docs panel for spacing system</div>