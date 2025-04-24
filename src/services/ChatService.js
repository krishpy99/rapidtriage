import { API_ENDPOINTS } from '../utils/config';

class ChatService {
  /**
   * Send text message to emergency endpoint
   * @param {string} message - The emergency message from the user
   * @param {Object} location - Object containing latitude and longitude
   * @returns {Promise} - Promise that resolves with formatted emergency response
   */
  async sendEmergencyMessage(message, location = null) {
    try {
      const payload = {
        text: message,
        timestamp: new Date().toISOString(),
      };

      // Add location data if available
      if (location && location.latitude && location.longitude) {
        payload.location = {
          latitude: location.latitude,
          longitude: location.longitude
        };
      }

      console.log('Sending emergency message with payload:', JSON.stringify(payload));
      console.log('API Endpoint:', API_ENDPOINTS.EMERGENCY_TEXT);

      const response = await fetch(API_ENDPOINTS.EMERGENCY_TEXT, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Accept': 'application/json'
        },
        body: JSON.stringify(payload)
      });

      // Check if response is ok before trying to parse JSON
      if (!response.ok) {
        // Try to get error message from response if possible
        let errorMessage;
        const contentType = response.headers.get('content-type');
        
        if (contentType && contentType.includes('application/json')) {
          // It's JSON, try to parse
          try {
            const errorData = await response.json();
            errorMessage = errorData.error || `Server error: ${response.status}`;
          } catch (e) {
            errorMessage = `Server error: ${response.status}`;
          }
        } else {
          // Not JSON, just get text
          try {
            const textResponse = await response.text();
            errorMessage = `Server error: ${response.status} - ${textResponse.substring(0, 100)}`;
          } catch (e) {
            errorMessage = `Server error: ${response.status}`;
          }
        }

        throw new Error(errorMessage);
      }
      
      // Safe to parse JSON now
      const contentType = response.headers.get('content-type');
      if (!contentType || !contentType.includes('application/json')) {
        console.warn(`Warning: Response is not JSON (${contentType})`);
        const textResponse = await response.text();
        console.log('Response text:', textResponse.substring(0, 200));
        throw new Error('Invalid response format: Expected JSON');
      }

      const data = await response.json();
      console.log('Emergency API response:', data);
      
      // Format the response for display
      const formattedResponse = this.formatEmergencyResponse(data);
      return formattedResponse;
    } catch (error) {
      console.error('Emergency message error:', error);
      throw error;
    }
  }

  /**
   * Format the emergency response for display
   * @param {Object} response - Raw emergency response data
   * @returns {Object} - Formatted response with message property
   */
  formatEmergencyResponse(response) {
    if (!response) return { message: "No response received from emergency services." };

    // Extract the main components
    const { emergency_id, code, summary, timestamp, tool_responses } = response;
    
    // Build a formatted message from the summary
    let formattedMessage = summary || "Emergency response received.";
    
    // Add information about tools that were called
    if (tool_responses && tool_responses.length > 0) {
      formattedMessage += '\n\n🚨 Actions taken:';
      
      tool_responses.forEach(tool => {
        const status = tool.success ? '✅' : '❌';
        formattedMessage += `\n${status} ${tool.tool_name}`;
      });
    }
    
    // Add emergency ID for reference
    if (emergency_id) {
      formattedMessage += `\n\nEmergency ID: ${emergency_id}`;
    }
    
    // Return the original response with an added message property for compatibility
    return {
      ...response,
      message: formattedMessage
    };
  }

  /**
   * Send emergency voice transcription to endpoint
   * @param {string} transcription - Transcribed audio content
   * @param {Object} location - Object containing latitude and longitude
   * @returns {Promise} - Promise that resolves with emergency response
   */
  async sendVoiceTranscription(transcription, location = null) {
    return this.sendEmergencyMessage(transcription, location);
  }
}

export default new ChatService();