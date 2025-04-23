# Backend Integration Guide for RapidTriage

This document outlines how to integrate this frontend with your friend's backend repository.

## API Endpoints

The frontend expects the following API endpoints:

1. `/convert-audio` - For speech-to-text conversion
- Method: POST
- Content-Type: multipart/form-data
- Request Body: FormData with 'audio' file
- Expected Response: 
  ```json
  {
    "success": true,
    "transcript": "Transcribed text here"
  }
  ```

2. `/analyze-symptoms` - For analyzing symptoms
- Method: POST
- Content-Type: application/json
- Request Body: 
  ```json
  {
    "symptoms": "User's symptom text",
    "duration": "less-than-24",
    "painLevel": 5
  }
  ```
- Expected Response:
  ```json
  {
    "requiresEmergencyCare": true,
    "recommendation": "Go to emergency room",
    "hospital": {
      "name": "General Hospital",
      "address": "123 Health St."
    }
  }
  ```

## Integration Steps

1. Update `src/utils/config.js` with the correct API base URL
2. Ensure the backend implements the required endpoints with the expected request/response formats
3. Test the integration by running the frontend with the backend
