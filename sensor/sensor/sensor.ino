#include <WiFi.h>
#include <WebServer.h>
#include <HTTPClient.h>
#include "config.h"  // Include configuration file

// Create a web server object
WebServer server(80);

// Pin where the magnetic sensor is connected
const int sensorPin = 4; // Adjust based on your setup
const int ledPin = 2;    // Pin for the LED (GPIO 2 on most ESP32 boards)

// Variable to hold the door status
String doorStatus = "Unknown";
String lastStatus = "Unknown";  // To track status changes

void setup() {
  // Initialize Serial Monitor
  Serial.begin(115200);

  // Set the sensor pin as input and LED pin as output
  pinMode(sensorPin, INPUT_PULLUP);
  pinMode(ledPin, OUTPUT);

  // Connect to Wi-Fi
  Serial.print("Connecting to Wi-Fi...");
  WiFi.begin(WIFI_SSID, WIFI_PASSWORD);

  while (WiFi.status() != WL_CONNECTED) {
    delay(500);
    Serial.print(".");
  }
  Serial.println("\nConnected! IP Address: ");
  Serial.println(WiFi.localIP());

  // Define the web server route
  server.on("/", handleRoot);

  // Start the web server
  server.begin();
  Serial.println("HTTP server started");
}

void loop() {
  // Check the door status
  int sensorValue = digitalRead(sensorPin);
  
  // Update door status
  if (sensorValue == LOW) {
    doorStatus = "Closed";
    digitalWrite(ledPin, LOW); // Turn LED off
  } else {
    doorStatus = "Open";
    digitalWrite(ledPin, HIGH); // Turn LED on
  }
  
  // Check if status changed from closed to open
  if (doorStatus == "Open" && lastStatus == "Closed") {
    Serial.println("Door just opened! Sending mail notification...");
    sendMailNotification();
  }
  
  // Print status only if it changed
  if (doorStatus != lastStatus) {
    Serial.println(doorStatus);
    lastStatus = doorStatus;  // Update last status
  }

  // Handle client requests
  server.handleClient();

  // Wait a short time before the next check
  delay(500);
}

void handleRoot() {
  // Create a webpage showing the door status
  String html = "<!DOCTYPE html><html>";
  html += "<head><meta http-equiv='refresh' content='1'><title>Door Status</title></head>";
  html += "<body><h1>Door Status</h1>";
  html += "<p>The door is currently: <strong>" + doorStatus + "</strong></p>";
  html += "</body></html>";

  // Send the webpage to the client
  server.send(200, "text/html", html);
}

void sendMailNotification() {
  // Check WiFi connection status
  if (WiFi.status() == WL_CONNECTED) {
    HTTPClient http;
    
    // Use the configured mail server URL
    http.begin(MAIL_SERVER);
    
    // Specify content-type header
    http.addHeader("Content-Type", "application/json");
    
    // Send the request
    int httpResponseCode = http.GET();
    
    if (httpResponseCode > 0) {
      String response = http.getString();
      Serial.println("HTTP Response code: " + String(httpResponseCode));
      Serial.println("Response: " + response);
    } else {
      Serial.print("Error on sending POST: ");
      Serial.println(httpResponseCode);
    }
    
    // Free resources
    http.end();
  } else {
    Serial.println("WiFi Disconnected, can't send notification");
  }
}