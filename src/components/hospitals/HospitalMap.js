import { Platform } from 'react-native';
import HospitalMapWeb from './platform/HospitalMapWeb';

// For web, use the web implementation
// For native, use the native implementation
let HospitalMap;

if (Platform.OS === 'web') {
  HospitalMap = HospitalMapWeb;
} else {
  // For native platforms, we'd import the native version
  // But for now, use the web version everywhere to avoid errors
  HospitalMap = HospitalMapWeb;
}

export default HospitalMap;
