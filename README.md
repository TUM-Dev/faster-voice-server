# faster-voice-service

Faster Voice Service is an attempt at implementing a transcription service
for TUM-Live/gocast that uses the faster-voice-server API provided by huggingface, 
to convert the audio of lectures into text. Furthermore, it aims at reducing
hallucinations, by running VAD (Voice Activity Detection) on the audio prior to 
transcription.
