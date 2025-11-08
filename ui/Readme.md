# Real Estate Data Visualization UI

## Overview

Web-based user interface for visualizing real estate market data across France, featuring interactive maps, statistical analysis, and multi-language support.

## Directory Structure

```text
ui/
├── frontend/          # React frontend application
├── index.html         # Entry HTML file
├── package.json       # Project dependencies and scripts
├── vite.config.ts     # Vite build configuration
└── jsdoc.config.json  # Documentation generator config
```

## Features

- Interactive map visualization
- Real estate price statistics
- Multi-level geographic analysis (City, Department, Region)
- Geocoding and address search
- Year-based filtering (2016-2023)
- Dark/Light theme support
- Internationalization (English and French)
- Responsive design

## Technology Stack

- React 18+
- Redux for state management
- Material-UI components
- Leaflet for maps
- i18next for translations
- Vite build system

## Getting Started

### Prerequisites

- Node.js 16 or higher
- npm or yarn package manager

### Installation

```bash
cd ui
npm install
```

### Development

Start the development server:

```bash
npm run dev
```

### Production Build

Create a production build:

```bash
npm run build
```

## Project Structure

### Core Components

- `App.jsx` - Main application container
- `MapViewer.jsx` - Interactive map component
- `Menubar.jsx` - Navigation and control bar
- Statistical components:
  - `CityStat.jsx`
  - `DepartmentStat.jsx`
  - `RegionStat.jsx`

### Services

- `poi_service.js` - Points of Interest API service
- `maputils.js` - Map utility functions

### State Management

- `store/index.js` - Redux store configuration
- `store/uiparamSlice.js` - UI state management

### Internationalization

Located in `frontend/locales/`:

- English (en, en-US)
- French (fr, fr-FR)

## Scripts

- `npm run dev` - Start development server
- `npm run build` - Create production build
- `npm run preview` - Preview production build
- `npm run docs` - Generate documentation

## Browser Support

- Chrome (latest)
- Firefox (latest)
- Safari (latest)
- Edge (latest)

## Development Guidelines

1. Follow React best practices
2. Use TypeScript for new components
3. Add JSDoc documentation
4. Include translations for new text
5. Test on both light and dark themes

## Contributing

1. Fork the repository
2. Create a feature branch
3. Commit changes
4. Push to the branch
5. Create a Pull Request

## License

MIT License
