import React, { useState, useRef, useEffect } from 'react';
import { 
  View, 
  Text, 
  StyleSheet, 
  SafeAreaView, 
  TextInput, 
  TouchableOpacity, 
  ScrollView,
  KeyboardAvoidingView,
  Platform,
  FlatList,
  Alert,
  Animated
} from 'react-native';
import VoiceRecorder from '../components/VoiceRecorder';
import LocationService from '../services/LocationService';
import ChatService from '../services/ChatService';

// Typing animation component to show when the system is processing
const TypingIndicator = () => {
  // Create three animated values for the dots
  const dot1Opacity = useRef(new Animated.Value(0.3)).current;
  const dot2Opacity = useRef(new Animated.Value(0.3)).current;
  const dot3Opacity = useRef(new Animated.Value(0.3)).current;
  
  // Animation sequence
  useEffect(() => {
    const animate = () => {
      // Reset values
      dot1Opacity.setValue(0.3);
      dot2Opacity.setValue(0.3);
      dot3Opacity.setValue(0.3);
      
      // Create animation sequence
      Animated.sequence([
        // First dot
        Animated.timing(dot1Opacity, {
          toValue: 1,
          duration: 300,
          useNativeDriver: true,
        }),
        // Second dot
        Animated.timing(dot2Opacity, {
          toValue: 1,
          duration: 300,
          useNativeDriver: true,
        }),
        // Third dot
        Animated.timing(dot3Opacity, {
          toValue: 1,
          duration: 300,
          useNativeDriver: true,
        }),
      ]).start(() => animate()); // Loop animation
    };
    
    animate();
    
    return () => {
      // Cleanup animations
      dot1Opacity.stopAnimation();
      dot2Opacity.stopAnimation();
      dot3Opacity.stopAnimation();
    };
  }, []);
  
  return (
    <View style={styles.typingIndicator}>
      <Animated.Text style={[styles.typingDot, { opacity: dot1Opacity }]}>•</Animated.Text>
      <Animated.Text style={[styles.typingDot, { opacity: dot2Opacity }]}>•</Animated.Text>
      <Animated.Text style={[styles.typingDot, { opacity: dot3Opacity }]}>•</Animated.Text>
    </View>
  );
};

const EmergencyScreen = ({ navigation }) => {
  const [messages, setMessages] = useState([]);
  const [inputText, setInputText] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [locationAvailable, setLocationAvailable] = useState(false);
  const [inputMode, setInputMode] = useState('text'); // 'text' or 'voice'
  const scrollViewRef = useRef();

  useEffect(() => {
    // Check if location services are available
    checkLocationServices();
  }, []);

  // Check if location services are available
  const checkLocationServices = async () => {
    try {
      await LocationService.getCurrentLocation();
      setLocationAvailable(true);
    } catch (error) {
      console.warn('Location services not available:', error);
      setLocationAvailable(false);
      // Only show the alert if it's not a user dismissal (permission denied would have its own UI)
      if (!error.message.includes('denied') && !error.message.includes('permission')) {
        Alert.alert(
          "Location Unavailable",
          "Your location can't be determined. Emergency services may be limited.",
          [{ text: "OK" }]
        );
      }
    }
  };

  // Safe get location with fallbacks
  const safeGetLocation = async () => {
    try {
      const location = await LocationService.getCurrentLocation();
      return {
        latitude: location.latitude,
        longitude: location.longitude
      };
    } catch (error) {
      console.warn('Could not get location:', error);
      // Return null if location is not available
      return null;
    }
  };

  // Switch between text and voice input modes
  const toggleInputMode = () => {
    setInputMode(prevMode => prevMode === 'text' ? 'voice' : 'text');
  };

  // Handle text submission
  const handleSendText = async () => {
    if (!inputText.trim()) return;
    
    // Add user message to chat
    const userMessage = {
      id: Date.now().toString(),
      text: inputText,
      sender: 'user'
    };
    
    setMessages(prevMessages => [...prevMessages, userMessage]);
    const textToSend = inputText;
    setInputText('');
    
    try {
      setIsLoading(true);
      
      // Get current location
      const locationData = await safeGetLocation();
      
      if (!locationData && !locationAvailable) {
        // If we've never been able to get location, inform the user
        const warningMessage = {
          id: (Date.now() + 1).toString(),
          text: "⚠️ Your location couldn't be determined. This may affect emergency response.",
          sender: 'system'
        };
        setMessages(prevMessages => [...prevMessages, warningMessage]);
      }
      
      // Send text to backend with location data
      const result = await ChatService.sendEmergencyMessage(textToSend, locationData);
      
      if (result && result.message) {
        // Add AI response to chat
        const aiMessage = {
          id: (Date.now() + 1).toString(),
          text: result.message,
          sender: 'ai'
        };
        setMessages(prevMessages => [...prevMessages, aiMessage]);
      } else {
        throw new Error(result?.error || 'Failed to process your message');
      }
    } catch (error) {
      console.error('Error sending text:', error);
      // Add error message with more specific information
      const errorMessage = {
        id: (Date.now() + 1).toString(),
        text: `Sorry, something went wrong: ${error.message || 'Unknown error'}. Please try again.`,
        sender: 'system'
      };
      setMessages(prevMessages => [...prevMessages, errorMessage]);
    } finally {
      setIsLoading(false);
    }
  };

  // Handle audio transcription reception
  const handleAudioProcessed = async (transcript) => {
    if (!transcript || transcript.trim() === '') {
      return; // Don't process empty transcripts
    }
    
    // Add user message to chat
    const aiResponse = {
      id: Date.now().toString(),
      text: transcript,
      sender: 'ai'
    };
    
    setMessages(prevMessages => [...prevMessages, aiResponse]);
    
    try {
      // Show loading alert
      Alert.alert(
        "Processing",
        "Processing your audio request...",
        [],
        { cancelable: false }
      );
      
      setIsLoading(true);
      
      // Get current location
      const locationData = await safeGetLocation();
      
      // Send transcript to backend with location data
      const result = await ChatService.sendVoiceTranscription(transcript, locationData);
      
      // Dismiss loading alert (by triggering another alert)
      if (result && result.message) {
        // Show response in an alert instead of in the chat
        Alert.alert(
          "Emergency Response",
          result.message,
          [{ text: "OK" }]
        );
      } else {
        throw new Error(result?.error || 'Failed to process your message');
      }
    } catch (error) {
      console.error('Error sending transcript:', error);
      // Show error as an alert
      Alert.alert(
        "Error",
        `Sorry, something went wrong: ${error.message || 'Unknown error'}. Please try again.`,
        [{ text: "OK" }]
      );
    } finally {
      setIsLoading(false);
    }
  };

  const renderChatMessage = ({ item }) => (
    <View style={[
      styles.messageContainer, 
      item.sender === 'user' ? styles.userMessage : 
      item.sender === 'ai' ? styles.aiMessage : styles.systemMessage
    ]}>
      <Text style={[
        styles.messageText,
        item.sender === 'system' && styles.systemMessageText
      ]}>{item.text}</Text>
    </View>
  );

  return (
    <SafeAreaView style={styles.container}>
      <View style={styles.header}>
        <TouchableOpacity 
          style={styles.backButton}
          onPress={() => navigation.goBack()}
        >
          <Text style={styles.backButtonText}>← Back</Text>
        </TouchableOpacity>
        <Text style={styles.headerTitle}>Emergency Assistance</Text>
      </View>
      
      <KeyboardAvoidingView 
        behavior={Platform.OS === "ios" ? "padding" : "height"}
        style={styles.content}
        keyboardVerticalOffset={0}
      >
        <View style={styles.chatContainer}>
          {messages.length === 0 ? (
            <View style={styles.emptyChat}>
              <Text style={styles.emptyChatText}>
                Describe your emergency situation by voice or text. Help is on the way.
              </Text>
              {!locationAvailable && (
                <Text style={styles.locationWarningText}>
                  ⚠️ Location services are not available. This may affect emergency response.
                </Text>
              )}
            </View>
          ) : (
            <>
              <FlatList
                data={messages}
                renderItem={renderChatMessage}
                keyExtractor={item => item.id}
                contentContainerStyle={styles.messagesList}
                ref={scrollViewRef}
                onContentSizeChange={() => scrollViewRef.current?.scrollToEnd({ animated: true })}
              />
              {isLoading && (
                <View style={styles.typingContainer}>
                  <TypingIndicator />
                </View>
              )}
            </>
          )}
        </View>
        
        <View style={styles.inputContainer}>
          {/* Input mode toggle button */}
          <TouchableOpacity 
            style={styles.toggleButton} 
            onPress={toggleInputMode}
          >
            <Text style={styles.toggleButtonText}>
              Switch to {inputMode === 'text' ? 'Voice' : 'Text'} Input
            </Text>
          </TouchableOpacity>
          
          {/* Show text input or voice recorder based on the input mode */}
          {inputMode === 'text' ? (
            <View style={styles.textInputContainer}>
              <TextInput
                style={styles.input}
                value={inputText}
                onChangeText={setInputText}
                placeholder="Type your emergency..."
                placeholderTextColor="#999"
                returnKeyType="send"
                multiline
                onSubmitEditing={handleSendText}
              />
              <TouchableOpacity
                style={[styles.sendButton, (!inputText.trim() || isLoading) && styles.sendButtonDisabled]}
                onPress={handleSendText}
                disabled={!inputText.trim() || isLoading}
              >
                <Text style={styles.sendButtonText}>Send</Text>
              </TouchableOpacity>
            </View>
          ) : (
            <View style={styles.recorderContainer}>
              <VoiceRecorder onAudioProcessed={handleAudioProcessed} />
            </View>
          )}
        </View>
      </KeyboardAvoidingView>
    </SafeAreaView>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#F5F5F5',
  },
  header: {
    flexDirection: 'row',
    alignItems: 'center',
    padding: 16,
    borderBottomWidth: 1,
    borderBottomColor: '#DDDDDD',
    backgroundColor: '#D32F2F',
  },
  backButton: {
    paddingRight: 16,
  },
  backButtonText: {
    fontSize: 16,
    fontWeight: '600',
    color: 'white',
  },
  headerTitle: {
    fontSize: 18,
    fontWeight: 'bold',
    color: 'white',
  },
  content: {
    flex: 1,
    display: 'flex',
    flexDirection: 'column',
  },
  chatContainer: {
    flex: 1,
    padding: 10,
  },
  emptyChat: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    padding: 20,
  },
  emptyChatText: {
    textAlign: 'center',
    fontSize: 16,
    color: '#666',
    lineHeight: 24,
    marginBottom: 10,
  },
  locationWarningText: {
    textAlign: 'center',
    fontSize: 14,
    color: '#D32F2F',
    marginTop: 10,
  },
  messagesList: {
    paddingVertical: 10,
  },
  messageContainer: {
    maxWidth: '85%',
    padding: 12,
    borderRadius: 16,
    marginVertical: 5,
  },
  userMessage: {
    backgroundColor: '#DCF8C5',
    alignSelf: 'flex-end',
    borderBottomRightRadius: 4,
  },
  aiMessage: {
    backgroundColor: 'white',
    alignSelf: 'flex-start',
    borderBottomLeftRadius: 4,
  },
  systemMessage: {
    backgroundColor: '#FFF3CD',
    alignSelf: 'center',
    borderRadius: 16,
    marginVertical: 8,
    width: '90%',
  },
  messageText: {
    fontSize: 16,
    color: '#333',
  },
  systemMessageText: {
    color: '#856404',
  },
  inputContainer: {
    padding: 0,
    paddingHorizontal: 10,
    paddingTop: 5,
    paddingBottom: 10,
    borderTopWidth: 1,
    borderTopColor: '#DDDDDD',
    backgroundColor: 'white',
  },
  textInputContainer: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  input: {
    flex: 1,
    backgroundColor: '#F0F0F0',
    borderRadius: 20,
    paddingHorizontal: 15,
    paddingVertical: 10,
    maxHeight: 100,
    fontSize: 16,
  },
  sendButton: {
    marginLeft: 10,
    backgroundColor: '#0A2F52',
    borderRadius: 20,
    paddingVertical: 10,
    paddingHorizontal: 15,
    justifyContent: 'center',
    alignItems: 'center',
  },
  sendButtonDisabled: {
    backgroundColor: '#CCCCCC',
  },
  sendButtonText: {
    color: 'white',
    fontWeight: '600',
  },
  recorderContainer: {
    marginTop: 0,
  },
  toggleButton: {
    backgroundColor: '#0A2F52',
    paddingVertical: 6,
    paddingHorizontal: 15,
    borderRadius: 20,
    alignSelf: 'center',
    marginBottom: 5,
    marginTop: 0,
  },
  toggleButtonText: {
    color: 'white',
    fontWeight: '600',
  },
  typingContainer: {
    padding: 8,
    marginLeft: 10,
    alignSelf: 'flex-start',
  },
  typingIndicator: {
    backgroundColor: 'white',
    borderRadius: 16,
    padding: 10,
    paddingHorizontal: 14,
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    borderBottomLeftRadius: 4,
  },
  typingDot: {
    fontSize: 24,
    marginHorizontal: 2,
    color: '#0A2F52',
  },
});

export default EmergencyScreen;