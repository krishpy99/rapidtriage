# RapidTriage App

A cross-platform medical triage application with voice input and nearby hospitals functionality.

## Features

- Symptom Assessment
  - Text and voice input for symptoms
  - Duration selection
  - Pain level rating
  - Analysis results

- Voice Recording
  - Record symptom descriptions
  - Automatic transcription to text
  - Visual recording feedback

- Nearby Hospitals
  - Find hospitals close to your location
  - Sort by rating and distance
  - View hospitals on a map
  - Get directions to hospital
  - Call hospital or emergency services

## Installation

1. Install dependencies
2. Configure API Keys
- Open `src/utils/config.js`
- Add your Google Places API key

3. Start the app
## Development

This project uses:
- React Native with Expo
- React Navigation for routing
- Expo AV for audio recording
- React Native Maps for mapping
- Google Places API for nearby hospitals

## Integration with Backend

To integrate with your friend's backend:

1. Update the API endpoints in `src/utils/config.js`
2. Replace mock API calls in components with actual implementation
3. Add any additional authentication needed
