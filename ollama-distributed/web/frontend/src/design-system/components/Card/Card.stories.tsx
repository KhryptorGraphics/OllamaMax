import type { Meta, StoryObj } from '@storybook/react'
import { Card, CardHeader, CardContent, CardFooter, CardTitle, CardDescription, CardImage, CardActions } from './Card'
import { Button } from '../Button/Button'
import { Badge } from '../Badge/Badge'
import { Heart, Share, MoreHorizontal, Star, Calendar, MapPin, User } from 'lucide-react'
import { useState } from 'react'

const meta: Meta<typeof Card> = {
  title: 'Design System/Card',
  component: Card,
  parameters: {
    docs: {
      description: {
        component: 'A flexible card component for grouping related content. Supports multiple variants, interactive states, and composable sub-components for headers, content, and footers.'
      }
    }
  },
  argTypes: {
    variant: {
      control: 'select',
      options: ['default', 'elevated', 'outlined', 'filled', 'interactive'],
      description: 'Visual style variant of the card'
    },
    padding: {
      control: 'select',
      options: ['none', 'sm', 'md', 'lg', 'xl'],
      description: 'Padding size for the card content'
    },
    interactive: {
      control: 'boolean',
      description: 'Whether the card is clickable'
    }
  },
  tags: ['autodocs']
}

export default meta
type Story = StoryObj<typeof Card>

// Default story
export const Default: Story = {
  args: {
    children: 'This is a basic card with default styling.'
  }
}

// Card variants
export const Variants: Story = {
  render: () => (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      <Card variant="default">
        <CardHeader>
          <CardTitle>Default Card</CardTitle>
          <CardDescription>
            Standard card with subtle shadow
          </CardDescription>
        </CardHeader>
        <CardContent>
          This is the default card variant with a subtle border and shadow.
        </CardContent>
      </Card>
      
      <Card variant="elevated">
        <CardHeader>
          <CardTitle>Elevated Card</CardTitle>
          <CardDescription>
            Card with enhanced shadow
          </CardDescription>
        </CardHeader>
        <CardContent>
          This card has a more prominent shadow for emphasis.
        </CardContent>
      </Card>
      
      <Card variant="outlined">
        <CardHeader>
          <CardTitle>Outlined Card</CardTitle>
          <CardDescription>
            Card with bold border
          </CardDescription>
        </CardHeader>
        <CardContent>
          This card has a thicker border and no shadow.
        </CardContent>
      </Card>
      
      <Card variant="filled">
        <CardHeader>
          <CardTitle>Filled Card</CardTitle>
          <CardDescription>
            Card with background fill
          </CardDescription>
        </CardHeader>
        <CardContent>
          This card has a filled background color.
        </CardContent>
      </Card>
      
      <Card variant="interactive" onCardClick={() => alert('Card clicked!')}>
        <CardHeader>
          <CardTitle>Interactive Card</CardTitle>
          <CardDescription>
            Clickable card with hover effects
          </CardDescription>
        </CardHeader>
        <CardContent>
          This card is clickable and has hover effects.
        </CardContent>
      </Card>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Different card variants for various use cases and visual emphasis.'
      }
    }
  }
}

// Padding variations
export const PaddingVariations: Story = {
  render: () => (
    <div className="space-y-4">
      <Card padding="none">
        <div className="p-2 bg-muted/50 m-2 rounded">
          <CardTitle size="sm">No Padding</CardTitle>
          <CardDescription>Card with no internal padding - content manages its own spacing</CardDescription>
        </div>
      </Card>
      
      <Card padding="sm">
        <CardTitle size="sm">Small Padding</CardTitle>
        <CardDescription>Compact spacing for dense layouts</CardDescription>
      </Card>
      
      <Card padding="md">
        <CardTitle>Medium Padding</CardTitle>
        <CardDescription>Standard spacing for most use cases</CardDescription>
      </Card>
      
      <Card padding="lg">
        <CardTitle>Large Padding</CardTitle>
        <CardDescription>Generous spacing for important content</CardDescription>
      </Card>
      
      <Card padding="xl">
        <CardTitle size="lg">Extra Large Padding</CardTitle>
        <CardDescription>Maximum spacing for hero content</CardDescription>
      </Card>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Different padding sizes for various content densities.'
      }
    }
  }
}

// Compound components
export const CompoundComponents: Story = {
  render: () => (
    <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
      <Card>
        <CardHeader divided>
          <CardTitle>User Profile</CardTitle>
          <CardDescription>
            Manage your account settings and preferences
          </CardDescription>
        </CardHeader>
        
        <CardContent spacing="md">
          <div className="flex items-center space-x-3">
            <div className="w-12 h-12 bg-primary/10 rounded-full flex items-center justify-center">
              <User className="w-6 h-6 text-primary" />
            </div>
            <div>
              <p className="font-medium">John Doe</p>
              <p className="text-sm text-muted-foreground">john@example.com</p>
            </div>
          </div>
          
          <div className="space-y-2">
            <div className="flex justify-between text-sm">
              <span>Account Type</span>
              <Badge variant="success">Premium</Badge>
            </div>
            <div className="flex justify-between text-sm">
              <span>Member Since</span>
              <span className="text-muted-foreground">Jan 2023</span>
            </div>
          </div>
        </CardContent>
        
        <CardFooter divided justify="between">
          <Button variant="outline" size="sm">
            Edit Profile
          </Button>
          <Button size="sm">
            View Details
          </Button>
        </CardFooter>
      </Card>
      
      <Card>
        <CardHeader>
          <div className="flex items-start justify-between">
            <div>
              <CardTitle>Project Status</CardTitle>
              <CardDescription>
                Current progress and next steps
              </CardDescription>
            </div>
            <Badge variant="warning">In Progress</Badge>
          </div>
        </CardHeader>
        
        <CardContent>
          <div className="space-y-4">
            <div>
              <div className="flex justify-between text-sm mb-1">
                <span>Progress</span>
                <span>75%</span>
              </div>
              <div className="w-full bg-muted rounded-full h-2">
                <div className="bg-primary h-2 rounded-full w-3/4"></div>
              </div>
            </div>
            
            <div className="grid grid-cols-2 gap-4 text-sm">
              <div>
                <p className="text-muted-foreground">Due Date</p>
                <p className="font-medium flex items-center">
                  <Calendar className="w-3 h-3 mr-1" />
                  Dec 15, 2024
                </p>
              </div>
              <div>
                <p className="text-muted-foreground">Team Size</p>
                <p className="font-medium">5 members</p>
              </div>
            </div>
          </div>
        </CardContent>
        
        <CardActions spacing="sm">
          <Button variant="outline" size="sm">
            View Details
          </Button>
          <Button size="sm">
            Update Status
          </Button>
        </CardActions>
      </Card>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Using compound components for structured card layouts with headers, content, and footers.'
      }
    }
  }
}

// Interactive cards
export const InteractiveCards: Story = {
  render: () => {
    const [selectedCard, setSelectedCard] = useState<number | null>(null)
    const [favoriteCards, setFavoriteCards] = useState<Set<number>>(new Set())

    const toggleFavorite = (id: number) => {
      const newFavorites = new Set(favoriteCards)
      if (newFavorites.has(id)) {
        newFavorites.delete(id)
      } else {
        newFavorites.add(id)
      }
      setFavoriteCards(newFavorites)
    }

    const cards = [
      {
        id: 1,
        title: 'Product A',
        description: 'High-quality product with excellent features',
        price: '$99',
        image: 'https://images.unsplash.com/photo-1526170375885-4d8ecf77b99f?w=300&h=200&fit=crop'
      },
      {
        id: 2,
        title: 'Product B',
        description: 'Innovative solution for modern needs',
        price: '$149',
        image: 'https://images.unsplash.com/photo-1560472354-b33ff0c44a43?w=300&h=200&fit=crop'
      },
      {
        id: 3,
        title: 'Product C',
        description: 'Premium quality with outstanding performance',
        price: '$199',
        image: 'https://images.unsplash.com/photo-1542291026-7eec264c27ff?w=300&h=200&fit=crop'
      }
    ]

    return (
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {cards.map(card => (
          <Card
            key={card.id}
            variant={selectedCard === card.id ? 'elevated' : 'interactive'}
            padding="none"
            interactive
            onCardClick={() => setSelectedCard(card.id)}
            className={selectedCard === card.id ? 'ring-2 ring-primary' : ''}
          >
            <CardImage
              src={card.image}
              alt={card.title}
              aspectRatio="video"
              position="top"
            />
            
            <div className="p-4">
              <CardHeader>
                <div className="flex items-start justify-between">
                  <div>
                    <CardTitle size="md">{card.title}</CardTitle>
                    <CardDescription>{card.description}</CardDescription>
                  </div>
                  <button
                    onClick={(e) => {
                      e.stopPropagation()
                      toggleFavorite(card.id)
                    }}
                    className="p-1 rounded-full hover:bg-muted transition-colors"
                  >
                    <Heart 
                      className={`w-4 h-4 ${
                        favoriteCards.has(card.id) 
                          ? 'fill-red-500 text-red-500' 
                          : 'text-muted-foreground'
                      }`}
                    />
                  </button>
                </div>
              </CardHeader>
              
              <CardContent>
                <div className="flex items-center justify-between">
                  <span className="text-lg font-bold">{card.price}</span>
                  <div className="flex items-center space-x-1 text-sm text-muted-foreground">
                    <Star className="w-3 h-3 fill-yellow-400 text-yellow-400" />
                    <span>4.5</span>
                  </div>
                </div>
              </CardContent>
              
              <CardActions>
                <Button 
                  size="sm" 
                  variant="outline"
                  onClick={(e) => e.stopPropagation()}
                >
                  Add to Cart
                </Button>
                <Button 
                  size="sm"
                  onClick={(e) => e.stopPropagation()}
                >
                  Buy Now
                </Button>
              </CardActions>
            </div>
          </Card>
        ))}
      </div>
    )
  },
  parameters: {
    docs: {
      description: {
        story: 'Interactive cards with click handlers, selection states, and nested interactive elements.'
      }
    }
  }
}

// Content cards
export const ContentCards: Story = {
  render: () => (
    <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
      <Card>
        <CardImage
          src="https://images.unsplash.com/photo-1557804506-669a67965ba0?w=400&h=250&fit=crop"
          alt="Article thumbnail"
          aspectRatio="video"
        />
        
        <CardHeader>
          <div className="flex items-center space-x-2 text-sm text-muted-foreground mb-2">
            <Badge variant="secondary" size="sm">Technology</Badge>
            <span>•</span>
            <span>5 min read</span>
          </div>
          <CardTitle>The Future of Web Development</CardTitle>
          <CardDescription>
            Exploring the latest trends and technologies shaping the future of web development.
          </CardDescription>
        </CardHeader>
        
        <CardContent>
          <div className="flex items-center space-x-2 text-sm text-muted-foreground">
            <Calendar className="w-3 h-3" />
            <span>Published on Dec 1, 2024</span>
          </div>
        </CardContent>
        
        <CardFooter>
          <div className="flex items-center justify-between w-full">
            <div className="flex items-center space-x-2">
              <div className="w-6 h-6 bg-primary/10 rounded-full flex items-center justify-center">
                <User className="w-3 h-3" />
              </div>
              <span className="text-sm">John Author</span>
            </div>
            
            <div className="flex items-center space-x-1">
              <Button variant="ghost" size="sm">
                <Heart className="w-4 h-4" />
              </Button>
              <Button variant="ghost" size="sm">
                <Share className="w-4 h-4" />
              </Button>
              <Button variant="ghost" size="sm">
                <MoreHorizontal className="w-4 h-4" />
              </Button>
            </div>
          </div>
        </CardFooter>
      </Card>
      
      <Card variant="outlined">
        <CardHeader>
          <CardTitle>Event Information</CardTitle>
          <CardDescription>
            Join us for an exciting tech conference
          </CardDescription>
        </CardHeader>
        
        <CardContent spacing="md">
          <div className="space-y-3">
            <div className="flex items-center space-x-2">
              <Calendar className="w-4 h-4 text-muted-foreground" />
              <span className="text-sm">December 15, 2024 at 9:00 AM</span>
            </div>
            
            <div className="flex items-center space-x-2">
              <MapPin className="w-4 h-4 text-muted-foreground" />
              <span className="text-sm">San Francisco Convention Center</span>
            </div>
            
            <div className="p-3 bg-muted/50 rounded-md">
              <p className="text-sm">
                Early bird pricing ends in 5 days! Get your tickets now and save 20%.
              </p>
            </div>
          </div>
        </CardContent>
        
        <CardActions justify="between">
          <span className="text-sm text-muted-foreground">Starting at $99</span>
          <div className="flex space-x-2">
            <Button variant="outline" size="sm">
              Learn More
            </Button>
            <Button size="sm">
              Register Now
            </Button>
          </div>
        </CardActions>
      </Card>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Content-rich cards for articles, events, and other informational content.'
      }
    }
  }
}

// Accessibility demonstration
export const AccessibilityDemo: Story = {
  render: () => (
    <div className="space-y-6">
      <div>
        <h3 className="text-sm font-medium mb-3">Keyboard Navigation</h3>
        <p className="text-sm text-muted-foreground mb-4">
          Interactive cards can be focused and activated with keyboard.
        </p>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <Card interactive onCardClick={() => alert('First card clicked!')}>
            <CardHeader>
              <CardTitle>First Card</CardTitle>
              <CardDescription>Press Enter or Space to activate</CardDescription>
            </CardHeader>
          </Card>
          
          <Card interactive onCardClick={() => alert('Second card clicked!')}>
            <CardHeader>
              <CardTitle>Second Card</CardTitle>
              <CardDescription>Tab to navigate between cards</CardDescription>
            </CardHeader>
          </Card>
        </div>
      </div>
      
      <div>
        <h3 className="text-sm font-medium mb-3">Semantic Structure</h3>
        <p className="text-sm text-muted-foreground mb-4">
          Cards use proper heading hierarchy and semantic markup.
        </p>
        <Card>
          <CardHeader>
            <CardTitle level={2}>Accessible Card</CardTitle>
            <CardDescription>
              This card uses proper semantic HTML structure for screen readers.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <p>Content is properly structured with headings and descriptions.</p>
          </CardContent>
        </Card>
      </div>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Accessibility features including keyboard navigation and semantic structure.'
      }
    }
  }
}

// Interactive playground
export const Playground: Story = {
  args: {
    variant: 'default',
    padding: 'md',
    interactive: false,
    children: 'Card content goes here. You can customize the variant, padding, and interactive behavior using the controls below.'
  },
  parameters: {
    docs: {
      description: {
        story: 'Interactive playground to test different card configurations.'
      }
    }
  }
}