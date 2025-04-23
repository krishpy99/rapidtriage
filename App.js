import React from 'react';
import { NavigationContainer } from '@react-navigation/native';
import { createStackNavigator } from '@react-navigation/stack';
import { StatusBar } from 'react-native';

// Import screens
import HomeScreen from './src/screens/HomeScreen';
import AssessmentScreen from './src/screens/AssessmentScreen';
import EmergencyResultsScreen from './src/screens/EmergencyResultsScreen';
import HospitalFinderScreen from './src/screens/HospitalFinderScreen';

// Create stack navigator
const Stack = createStackNavigator();

export default function App() {
  return (
    <NavigationContainer>
      <StatusBar barStyle="dark-content" backgroundColor="#E5E5E5" />
      <Stack.Navigator
        initialRouteName="Home"
        screenOptions={{
          headerShown: false,
          cardStyle: { backgroundColor: '#F5F5F5' }
        }}
      >
        <Stack.Screen name="Home" component={HomeScreen} />
        <Stack.Screen name="Assessment" component={AssessmentScreen} />
        <Stack.Screen name="EmergencyResults" component={EmergencyResultsScreen} />
        <Stack.Screen name="HospitalFinder" component={HospitalFinderScreen} />
      </Stack.Navigator>
    </NavigationContainer>
  );
}
