import { Platform } from 'react-native';
import * as Location from 'expo-location';

import { API_ENDPOINTS } from '../utils/config';

class LocationService {
  // Get current location with platform-specific implementation
  async getCurrentLocation() {
    try {
      // Request location permissions first
      let { status } = await Location.requestForegroundPermissionsAsync();
      if (status !== 'granted') {
        throw new Error('Location permission not granted');
      }

      // Get current position using expo-location
      const location = await Location.getCurrentPositionAsync({
        accuracy: Location.Accuracy.High,
      });
      
      return {
        latitude: location.coords.latitude,
        longitude: location.coords.longitude,
      };
    } catch (error) {
      console.error('Location error:', error);
      throw error;
    }
  }

  // Find nearby hospitals
  async findNearbyHospitals(userLocation, radius = 5000) {
    try {
      const { latitude, longitude } = userLocation;
      
      // API call to Google Places API
      const response = await fetch(
        `https://maps.googleapis.com/maps/api/place/nearbysearch/json?location=${latitude},${longitude}&radius=${radius}&type=hospital&key=${API_ENDPOINTS.GOOGLE_PLACES_API_KEY}`
      );
      
      const data = await response.json();
      
      if (data.status !== 'OK') {
        throw new Error(`API Error: ${data.status}`);
      }
      
      return data.results;
    } catch (error) {
      console.error('Find hospitals error:', error);
      throw error;
    }
  }

  // Calculate distance between two points using Haversine formula
  calculateDistance(lat1, lon1, lat2, lon2) {
    const R = 6371; // Radius of the earth in km
    const dLat = this.deg2rad(lat2 - lat1);
    const dLon = this.deg2rad(lon2 - lon1);
    const a = 
      Math.sin(dLat / 2) * Math.sin(dLat / 2) +
      Math.cos(this.deg2rad(lat1)) * Math.cos(this.deg2rad(lat2)) * 
      Math.sin(dLon / 2) * Math.sin(dLon / 2);
    const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));
    const distance = R * c; // Distance in km
    return distance;
  }
  
  // Convert degrees to radians
  deg2rad(deg) {
    return deg * (Math.PI / 180);
  }

  // Format distance for display
  formatDistance(distance) {
    return distance < 1 
      ? `${(distance * 1000).toFixed(0)} m` 
      : `${distance.toFixed(1)} km`;
  }
}

export default new LocationService();
