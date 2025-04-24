// Configuration for API endpoints
// Using API endpoints exclusively from environment variables
import { 
  API_BASE_URL, 
  GOOGLE_PLACES_API_KEY,
} from "@env";

// Use the base API URL from environment variable
const BASE_URL = API_BASE_URL || '';

export const API_ENDPOINTS = {
  // Health check endpoint (or construct from base URL if not specifically provided)
  HEALTH: `${BASE_URL}/health`,
  
  // Emergency endpoints
  EMERGENCY: `${BASE_URL}/emergency`,
  EMERGENCY_TEXT: `${BASE_URL}/emergency/text`,
  
  // Google Places API key from environment variables
  GOOGLE_PLACES_API_KEY: GOOGLE_PLACES_API_KEY || '',
};

export default API_ENDPOINTS;
