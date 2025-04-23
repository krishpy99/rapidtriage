import React from 'react';
import { 
  View, 
  Text, 
  TouchableOpacity, 
  StyleSheet, 
  SafeAreaView 
} from 'react-native';
import BrowserHeader from '../components/BrowserHeader';

const HomeScreen = ({ navigation }) => {
  // Make sure navigation is defined before using it
  const handleStartAssessment = () => {
    if (navigation) {
      navigation.navigate('Assessment');
    } else {
      console.error('Navigation is undefined');
    }
  };

  return (
    <SafeAreaView style={styles.container}>
      <BrowserHeader />
      
      <View style={styles.content}>
        <Text style={styles.title}>RapidTriage AI</Text>
        
        <Text style={styles.subtitle}>
          Get quick assessment of your symptoms and find out if you need emergency care.
        </Text>
        
        <TouchableOpacity
          style={styles.startButton}
          onPress={handleStartAssessment}
        >
          <Text style={styles.startButtonText}>Start Assessment</Text>
        </TouchableOpacity>
      </View>
      
      <Text style={styles.disclaimer}>
        For informational purposes only.{'\n'}
        Not a substitute for professional medical advice.
      </Text>
    </SafeAreaView>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#F5F5F5',
  },
  content: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    padding: 20,
  },
  title: {
    fontSize: 32,
    fontWeight: 'bold',
    color: '#333333',
    marginBottom: 20,
    textAlign: 'center',
  },
  subtitle: {
    fontSize: 18,
    color: '#555555',
    textAlign: 'center',
    marginBottom: 40,
    lineHeight: 24,
  },
  startButton: {
    backgroundColor: '#0A2F52',
    paddingVertical: 14,
    paddingHorizontal: 30,
    borderRadius: 25,
    marginBottom: 20,
  },
  startButtonText: {
    color: '#FFFFFF',
    fontSize: 18,
    fontWeight: 'bold',
  },
  disclaimer: {
    textAlign: 'center',
    color: '#777777',
    fontSize: 12,
    lineHeight: 18,
    marginBottom: 20,
  },
});

export default HomeScreen;
