import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:web_socket_channel/web_socket_channel.dart';

void main() {
  runApp(const MyApp());
}

class MyApp extends StatelessWidget {
  const MyApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Real-time Poll',
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(seedColor: Colors.deepPurple),
        useMaterial3: true,
      ),
      home: const PollPage(),
    );
  }
}

class PollPage extends StatefulWidget {
  const PollPage({super.key});

  @override
  State<PollPage> createState() => _PollPageState();
}

class _PollPageState extends State<PollPage> {
  late final WebSocketChannel _channel;

  @override
  void initState() {
    super.initState();
    final host = Uri.base.host;
    _channel = WebSocketChannel.connect(Uri.parse('ws://$host:8080/ws'));
  }

  @override
  void dispose() {
    _channel.sink.close();
    super.dispose();
  }

  // Функция для отправки голоса на сервер
  void _sendVote(int optionId) {
    final message = jsonEncode({'option_id': optionId});
    _channel.sink.add(message);
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Голосование в реальном времени'),
        backgroundColor: Theme.of(context).colorScheme.inversePrimary,
      ),
      body: Padding(
        padding: const EdgeInsets.all(16.0),
        child: StreamBuilder(
          stream: _channel.stream,
          builder: (context, snapshot) {
            if (snapshot.connectionState == ConnectionState.waiting) {
              return const Center(child: CircularProgressIndicator());
            }

            if (snapshot.hasError) {
              return Center(child: Text('Ошибка связи: ${snapshot.error}'));
            }

            if (snapshot.hasData) {
              final data = jsonDecode(snapshot.data as String);

              final question = data['question'] as String;
              final options = data['options'] as List<dynamic>;

              return Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    question,
                    style: const TextStyle(
                      fontSize: 22,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  const SizedBox(height: 24),

                  Expanded(
                    child: ListView.builder(
                      itemCount: options.length,
                      itemBuilder: (context, index) {
                        final option = options[index];
                        final int id = option['id'];
                        final stringText = option['text'] as String;
                        final int votes = option['votes'];

                        return Card(
                          margin: const EdgeInsets.symmetric(vertical: 8),
                          child: ListTile(
                            title: Text(
                              stringText,
                              style: const TextStyle(fontSize: 16),
                            ),
                            trailing: Container(
                              padding: const EdgeInsets.symmetric(
                                horizontal: 12,
                                vertical: 6,
                              ),
                              decoration: BoxDecoration(
                                color: Colors.deepPurple.withValues(alpha: 0.1),
                                borderRadius: BorderRadius.circular(12),
                              ),
                              child: Text(
                                '$votes голосов',
                                style: const TextStyle(
                                  fontWeight: FontWeight.bold,
                                  color: Colors.deepPurple,
                                ),
                              ),
                            ),
                            onTap: () => _sendVote(id),
                          ),
                        );
                      },
                    ),
                  ),
                ],
              );
            }

            return const Center(child: Text('Нет данных'));
          },
        ),
      ),
    );
  }
}
