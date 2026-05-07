-- Seed SEO Pages with all existing routes from sitemap-generator.ts

INSERT INTO seo_pages (slug, type, title, meta_description, meta_keywords, content_config, priority, changefreq) VALUES

-- === Core Pages ===
('home', 'core', 'Notly - Öğrenci Planlama ve Verimlilik Uygulaması',
 'YKS, üniversite ve lise öğrencileri için ücretsiz planlama uygulaması. Ders programı, görev takibi, alışkanlık yönetimi ve sınav takvimi.',
 ARRAY['öğrenci planlayıcı', 'ders çalışma programı', 'yks hazırlık'],
 '{"page_type": "home"}'::jsonb, 1.0, 'daily'),

('tasks', 'core', 'Görev Yönetimi - Notly',
 'Ödevlerinizi, projelerinizi ve günlük görevlerinizi kolayca takip edin.',
 ARRAY['görev yöneticisi', 'ödev takip', 'yapılacaklar listesi'],
 '{"page_type": "feature"}'::jsonb, 0.9, 'daily'),

('habits', 'core', 'Alışkanlık Takibi - Notly',
 'Pozitif alışkanlıklar edinin, kötü alışkanlıkları bırakın. Günlük streakler ve motivasyon sistemi.',
 ARRAY['alışkanlık takip', 'günlük rutin', 'hedef belirleme'],
 '{"page_type": "feature"}'::jsonb, 0.9, 'daily'),

('courses', 'core', 'Ders Programı Yönetimi - Notly',
 'Derslerinizi, sınav tarihlerini ve ders notlarınızı tek platformda yönetin.',
 ARRAY['ders programı', 'sınav takvimi', 'akademik takip'],
 '{"page_type": "feature"}'::jsonb, 0.8, 'weekly'),

('goals', 'core', 'Hedef Takibi - Notly',
 'Akademik ve kişisel hedeflerinizi belirleyin ve takip edin.',
 ARRAY['hedef belirleme', 'akademik hedefler', 'başarı takibi'],
 '{"page_type": "feature"}'::jsonb, 0.8, 'weekly'),

('blog', 'core', 'Blog - Notly',
 'Öğrenciler için verimlilik, çalışma teknikleri ve akademik başarı yazıları.',
 ARRAY['öğrenci blogları', 'çalışma teknikleri', 'akademik yazılar'],
 '{"page_type": "blog"}'::jsonb, 0.7, 'weekly'),

-- === Templates (Programmatic SEO) ===
('yks-calisma-programi', 'template',
 'YKS 2026-2027 Ders Çalışma Programı Şablonu',
 'TYT ve AYT sınavlarına hazırlık için haftalık ders çalışma programı. Ücretsiz kullan, indirmeden düzenle.',
 ARRAY['yks çalışma programı', 'tyt ayt plan', 'ders çalışma takvimi'],
 '{"template_type": "study_plan", "target_exam": "yks"}'::jsonb, 0.8, 'monthly'),

('vize-final-takip', 'template',
 'Vize Final Sınav Takip Şablonu - Üniversite',
 'Vize ve final sınavlarınızı takip edin. Sınav tarihleri, konu listesi, çalışma planı içerir.',
 ARRAY['vize final takip', 'sınav takvimi', 'dönem sonu hazırlık'],
 '{"template_type": "exam_tracker"}'::jsonb, 0.8, 'monthly'),

('tyt-ayt-planlama', 'template',
 'TYT AYT Planlama Şablonu - YKS Hazırlık',
 'TYT ve AYT konularını planlayın, eksik konuları takip edin.',
 ARRAY['tyt ayt planlama', 'yks konu takip', 'sınav hazırlık'],
 '{"template_type": "topic_planner"}'::jsonb, 0.8, 'monthly'),

('ders-programi', 'template',
 'Haftalık Ders Programı Şablonu',
 'Haftalık ders programınızı oluşturun, saat bazlı planlama yapın.',
 ARRAY['haftalık ders programı', 'saat programı', 'ders çizelgesi'],
 '{"template_type": "weekly_schedule"}'::jsonb, 0.7, 'monthly'),

('staj-gunlugu', 'template',
 'Staj Günlüğü Şablonu - Üniversite',
 'Staj sürecinizi kaydedin, günlük raporlar tutun.',
 ARRAY['staj günlüğü', 'staj raporu', 'mesleki gelişim'],
 '{"template_type": "internship_journal"}'::jsonb, 0.7, 'monthly'),

-- === YKS Calculators (High Volume) ===
('yks-puan-hesaplama', 'tool',
 'YKS Puan Hesaplama 2026 - Ücretsiz Online Hesaplayıcı',
 'TYT ve AYT netlerinizi girin, YKS puanınızı anında hesaplayın. 2026 katsayıları ile güncel hesaplama.',
 ARRAY['yks puan hesaplama', 'tyt ayt puan', 'üniversite taban puanı', 'yks 2026'],
 '{"tool_type": "yks_calculator", "calculator_mode": "combined"}'::jsonb, 0.95, 'monthly'),

('tyt-puan-hesaplama', 'tool',
 'TYT Puan Hesaplama 2026 - Net Hesaplayıcı',
 'TYT netlerinizle ham puanınızı hesaplayın. Türkçe, Matematik, Fen, Sosyal net girişi.',
 ARRAY['tyt puan hesaplama', 'tyt net hesaplama', 'tyt sıralama'],
 '{"tool_type": "tyt_calculator"}'::jsonb, 0.90, 'monthly'),

('ayt-puan-hesaplama', 'tool',
 'AYT Puan Hesaplama 2026 - Sayısal Sözel EA',
 'AYT Sayısal, Sözel, Eşit Ağırlık, Dil puanlarını hesaplayın.',
 ARRAY['ayt puan hesaplama', 'ayt sayısal', 'ayt sözel', 'ayt ea'],
 '{"tool_type": "ayt_calculator"}'::jsonb, 0.90, 'monthly'),

('yks-sayisal-puan', 'tool',
 'YKS Sayısal Puan Hesaplama - Mühendislik Tıp',
 'Sayısal alan YKS puanı ve sıralamanızı hesaplayın. Mühendislik, Tıp, Fen fakülteleri için.',
 ARRAY['yks sayısal puan', 'mühendislik puan hesaplama', 'tıp puanı'],
 '{"tool_type": "yks_calculator", "field": "sayisal"}'::jsonb, 0.90, 'monthly'),

('yks-sozel-puan', 'tool',
 'YKS Sözel Puan Hesaplama - Hukuk Edebiyat',
 'Sözel alan YKS puanı hesaplama. Hukuk, İktisat, İşletme, Edebiyat için.',
 ARRAY['yks sözel puan', 'hukuk puanı', 'edebiyat puan'],
 '{"tool_type": "yks_calculator", "field": "sozel"}'::jsonb, 0.90, 'monthly'),

('yks-ea-puan', 'tool',
 'YKS Eşit Ağırlık (EA) Puan Hesaplama',
 'EA puan hesaplama ve sıralama tahmini. İktisat, İşletme, İletişim için.',
 ARRAY['yks ea puan', 'eşit ağırlık', 'iktisat puanı'],
 '{"tool_type": "yks_calculator", "field": "ea"}'::jsonb, 0.90, 'monthly'),

('yks-siralama-hesaplama', 'tool',
 'YKS Sıralama Hesaplama 2026 - Geçmiş Yıl Karşılaştırma',
 'YKS puanınızla tahmini sıralamanızı ve geçen yıl verisiyle karşılaştırın.',
 ARRAY['yks sıralama hesaplama', 'yks sıralama tahmini', '2025 sıralama'],
 '{"tool_type": "yks_ranking_calculator"}'::jsonb, 0.95, 'monthly'),

('obp-yks-puan', 'tool',
 'OBP''li YKS Puan Hesaplama - Diploma Notu Ekle',
 'Okul başarı puanı (OBP) ekleyerek YKS puanınızı hesaplayın.',
 ARRAY['obp yks puan', 'diploma notu yks', 'okul puanı'],
 '{"tool_type": "yks_calculator", "include_obp": true}'::jsonb, 0.90, 'monthly'),

('obp-siz-yks', 'tool',
 'OBP''siz (Ham) YKS Puan Hesaplama - Mezun',
 'Mezunlar ve OBP olmadan YKS ham puan hesaplama.',
 ARRAY['obp siz yks', 'ham puan hesaplama', 'mezun yks'],
 '{"tool_type": "yks_calculator", "include_obp": false}'::jsonb, 0.85, 'monthly'),

('kac-net-kac-puan', 'tool',
 'Kaç Net Kaç Puan Getirir? - YKS Net-Puan Dönüşümü',
 'Her net sayısının karşılığı kaç puan olduğunu öğrenin.',
 ARRAY['kaç net kaç puan', 'yks net puan tablosu', 'net puan karşılığı'],
 '{"tool_type": "net_to_score_converter"}'::jsonb, 0.95, 'monthly'),

-- === GPA Calculators (High Retention) ===
('gpa-hesaplama', 'tool',
 'GPA / Not Ortalaması Hesaplama - AA BA BB Sistemi',
 'Üniversite not ortalamanızı hesaplayın. 4''lük ve 100''lük sistem desteği.',
 ARRAY['gpa hesaplama', 'not ortalaması', 'üniversite not'],
 '{"tool_type": "gpa_calculator", "systems": ["4.0", "100"]}'::jsonb, 0.90, 'monthly'),

('gano-hesaplama', 'tool',
 'GANO Hesaplama - Genel Akademik Not Ortalaması',
 'Tüm dönemlerinizin genel not ortalamasını hesaplayın.',
 ARRAY['gano hesaplama', 'genel not ortalaması', 'cumulative gpa'],
 '{"tool_type": "gpa_calculator", "gpa_type": "gano"}'::jsonb, 0.90, 'monthly'),

('yano-hesaplama', 'tool',
 'YANO Hesaplama - Yarıyıl Not Ortalaması',
 'Dönemlik akademik not ortalamanızı hesaplayın.',
 ARRAY['yano hesaplama', 'dönem ortalaması', 'semester gpa'],
 '{"tool_type": "gpa_calculator", "gpa_type": "yano"}'::jsonb, 0.85, 'monthly'),

('agno-hesaplama', 'tool',
 'AGNO Hesaplama - Ağırlıklı Genel Not Ortalaması',
 'Kredi ağırlıklı genel not ortalaması hesaplama.',
 ARRAY['agno hesaplama', 'ağırlıklı ortalama', 'weighted gpa'],
 '{"tool_type": "gpa_calculator", "gpa_type": "agno"}'::jsonb, 0.85, 'monthly'),

('4luk-100luk-cevirme', 'tool',
 '4''lük Sistemi 100''lük Sisteme Çevirme',
 '4.0 üzerinden notunuzu 100''lük sisteme dönüştürün.',
 ARRAY['4lük 100lük çevirme', 'gpa conversion', 'not dönüşümü'],
 '{"tool_type": "gpa_converter", "from": "4.0", "to": "100"}'::jsonb, 0.85, 'monthly'),

('100luk-4luk-cevirme', 'tool',
 '100''lük Sistemi 4''lük Sisteme Çevirme',
 '100 üzerinden notunuzu 4.0 sistemine dönüştürün.',
 ARRAY['100lük 4lük çevirme', 'not dönüşümü', 'gpa conversion'],
 '{"tool_type": "gpa_converter", "from": "100", "to": "4.0"}'::jsonb, 0.85, 'monthly'),

('harf-notu-hesaplama', 'tool',
 'Harf Notu Hesaplama - AA BA BB CB Nedir?',
 'Sayısal notunuzun harf karşılığını ve GPA değerini hesaplayın.',
 ARRAY['harf notu hesaplama', 'aa ba bb', 'letter grade'],
 '{"tool_type": "letter_grade_calculator"}'::jsonb, 0.85, 'monthly'),

-- University-Specific GPA
('itu-ortalama', 'tool',
 'İTÜ Ortalama Hesaplama - İstanbul Teknik Üniversitesi GANO',
 'İTÜ''ye özel not sistemi ile ortalama hesaplama.',
 ARRAY['itü ortalama', 'itü gano', 'istanbul teknik not'],
 '{"tool_type": "gpa_calculator", "university": "itu"}'::jsonb, 0.80, 'monthly'),

('odtu-gpa', 'tool',
 'ODTÜ GPA Calculator - Orta Doğu Teknik Üniversitesi',
 'ODTÜ harf notu sistemine göre GPA hesaplama.',
 ARRAY['odtü gpa', 'odtü ortalama', 'metu gpa'],
 '{"tool_type": "gpa_calculator", "university": "odtu"}'::jsonb, 0.80, 'monthly'),

('aof-puan', 'tool',
 'AÖF Puan Hesaplama - Açıköğretim Not Ortalaması',
 'Açıköğretim Fakültesi ortalama hesaplama sistemi.',
 ARRAY['aöf puan hesaplama', 'açıköğretim ortalama', 'aof gpa'],
 '{"tool_type": "gpa_calculator", "university": "aof"}'::jsonb, 0.80, 'monthly'),

-- === Vize-Final Calculators (Killer Feature) ===
('vize-final-hesaplama', 'tool',
 'Vize Final Hesaplama - Dersi Geçme Notu Hesaplayıcı',
 'Vize notunuzu girin, finalden kaç almanız gerektiğini hesaplayın.',
 ARRAY['vize final hesaplama', 'finalden kaç almam lazım', 'dersi geçme'],
 '{"tool_type": "vize_final_calculator", "weights": {"vize": 0.4, "final": 0.6}, "passing_grade": 50}'::jsonb, 0.95, 'monthly'),

('dersi-gecme-notu', 'tool',
 'Dersi Geçme Notu Hesaplama - Geçmek İçin Final Notu',
 'Dersi geçmek için finalden almanız gereken minimum notu hesaplayın.',
 ARRAY['dersi geçmek için', 'final notu hesaplama', 'geçme notu'],
 '{"tool_type": "vize_final_calculator", "mode": "passing_grade"}'::jsonb, 0.90, 'monthly'),

('finalden-kac-lazim', 'tool',
 'Finalden Kaç Almam Lazım? - Vize Notu ile Hesaplama',
 'Vize notunuza göre finalden kaç almanız gerektiğini anında öğrenin.',
 ARRAY['finalden kaç almam lazım', 'final hesaplama', 'gerekli final'],
 '{"tool_type": "vize_final_calculator", "mode": "required_final"}'::jsonb, 0.95, 'monthly'),

('bute-kalmamak', 'tool',
 'Büte Kalmamak İçin Kaç Lazım? - Büt Hesaplama',
 'Bütünlemeden kurtulmak için final notu hesaplama.',
 ARRAY['büte kalmamak', 'büt hesaplama', 'bütünleme notu'],
 '{"tool_type": "vize_final_calculator", "mode": "avoid_resit"}'::jsonb, 0.90, 'monthly'),

('vize-40-final-60', 'tool',
 'Vize %40 Final %60 Hesaplama - Standart Sistem',
 'Vize %40, Final %60 ağırlıklı not ortalaması hesaplama.',
 ARRAY['vize 40 final 60', 'not ağırlık hesaplama', '40 60 sistem'],
 '{"tool_type": "vize_final_calculator", "weights": {"vize": 0.4, "final": 0.6}}'::jsonb, 0.85, 'monthly'),

('vize-30-final-70', 'tool',
 'Vize %30 Final %70 Hesaplama',
 'Vize %30, Final %70 ağırlıklı not hesaplama sistemi.',
 ARRAY['vize 30 final 70', 'not ortalaması', '30 70 sistem'],
 '{"tool_type": "vize_final_calculator", "weights": {"vize": 0.3, "final": 0.7}}'::jsonb, 0.85, 'monthly'),

('can-egrisi', 'tool',
 'Çan Eğrisi Hesaplama - Bağıl Değerlendirme Sistemi',
 'Sınıf ortalamasına göre çan eğrisi ile harf notunuzu hesaplayın.',
 ARRAY['çan eğrisi', 'bağıl değerlendirme', 'curve grading'],
 '{"tool_type": "curve_grading_calculator"}'::jsonb, 0.75, 'monthly'),

-- === Other Tools ===
('pomodoro', 'tool',
 'Pomodoro Sayacı - Ücretsiz Çalışma Zamanlayıcı',
 '25 dakika odaklanma, 5 dakika mola. Pomodoro tekniği ile verimli çalışın.',
 ARRAY['pomodoro sayacı', 'çalışma zamanlayıcı', 'verimlilik timer'],
 '{"tool_type": "pomodoro_timer", "work_duration": 25, "break_duration": 5}'::jsonb, 0.80, 'monthly'),

('ders-programi-olustur', 'tool',
 'Ders Programı Oluştur - Otomatik Haftalık Plan',
 'Derslerinizi girin, otomatik haftalık ders programı oluşturun.',
 ARRAY['ders programı oluştur', 'otomatik program', 'haftalık çizelge'],
 '{"tool_type": "schedule_generator"}'::jsonb, 0.80, 'monthly'),

('haftalik-planlayici', 'tool',
 'Haftalık Planlayıcı - Görev ve Ders Organizasyonu',
 'Haftanızı planlayın, görevlerinizi ve derslerinizi organize edin.',
 ARRAY['haftalık planlayıcı', 'görev planlama', 'ders organizasyon'],
 '{"tool_type": "weekly_planner"}'::jsonb, 0.75, 'monthly'),

-- === Content/Guides ===
('vize-sinavina-nasil-hazirlanilir', 'guide',
 'Vize Sınavına Nasıl Hazırlanılır? - Etkili Çalışma Yöntemleri',
 'Vize sınavlarına hazırlık için kanıtlanmış çalışma stratejileri ve ipuçları.',
 ARRAY['vize hazırlık', 'sınav çalışma', 'vize stratejileri'],
 '{"guide_type": "exam_prep", "exam": "vize"}'::jsonb, 0.70, 'monthly'),

('final-haftasi-calisma-stratejileri', 'guide',
 'Final Haftası Çalışma Stratejileri - Son Hafta Planı',
 'Final haftasında verimli çalışma yöntemleri ve zaman yönetimi teknikleri.',
 ARRAY['final çalışma', 'son hafta planı', 'final stratejileri'],
 '{"guide_type": "exam_prep", "exam": "final"}'::jsonb, 0.70, 'monthly'),

('yks-son-30-gun-plani', 'guide',
 'YKS Son 30 Gün Planı - Sınav Öncesi Hazırlık',
 'YKS sınavına son 30 günde maksimum verimle nasıl hazırlanılır.',
 ARRAY['yks son 30 gün', 'yks hazırlık', 'sınav öncesi plan'],
 '{"guide_type": "exam_prep", "exam": "yks", "duration_days": 30}'::jsonb, 0.70, 'monthly'),

('universite-ders-notu-tutma', 'guide',
 'Üniversite Ders Notu Tutma Teknikleri - Cornell Yöntemi',
 'Etkili not tutma yöntemleri, Cornell sistemi ve dijital araçlar.',
 ARRAY['not tutma teknikleri', 'cornell yöntemi', 'ders notu'],
 '{"guide_type": "study_skills", "topic": "note_taking"}'::jsonb, 0.65, 'monthly'),

('but-sinavina-hazirlik', 'guide',
 'Büt Sınavına Hazırlık - Bütünlemeden Başarıyla Geçme',
 'Bütünleme sınavına kısa sürede nasıl hazırlanılır, etkili teknikler.',
 ARRAY['büt hazırlık', 'bütünleme çalışma', 'resit exam'],
 '{"guide_type": "exam_prep", "exam": "resit"}'::jsonb, 0.65, 'monthly'),

-- === Features ===
('internetsiz-calisma', 'feature',
 'İnternetsiz Çalışma - Offline Öğrenci Planlayıcı',
 'Notly uygulaması internet bağlantısı olmadan çalışır. Tüm verilere offline erişin.',
 ARRAY['offline uygulama', 'internetsiz not', 'çevrimdışı çalışma'],
 '{"feature_type": "offline_mode", "pwa": true}'::jsonb, 0.80, 'monthly');
