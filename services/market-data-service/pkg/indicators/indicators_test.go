package indicators

import (
    "testing"
    "time"

    "github.com/algo-trading/market-data-service/internal/models"
)

func createTestOHLCVData() []models.OHLCV {
    data := []models.OHLCV{
        {Time: time.Now(), Symbol: "TEST", Open: 100, High: 105, Low: 98, Close: 102, Volume: 1000, Timeframe: "1d"},
        {Time: time.Now(), Symbol: "TEST", Open: 102, High: 108, Low: 101, Close: 106, Volume: 1100, Timeframe: "1d"},
        {Time: time.Now(), Symbol: "TEST", Open: 106, High: 110, Low: 104, Close: 108, Volume: 1200, Timeframe: "1d"},
        {Time: time.Now(), Symbol: "TEST", Open: 108, High: 112, Low: 106, Close: 110, Volume: 1300, Timeframe: "1d"},
        {Time: time.Now(), Symbol: "TEST", Open: 110, High: 115, Low: 109, Close: 113, Volume: 1400, Timeframe: "1d"},
        {Time: time.Now(), Symbol: "TEST", Open: 113, High: 117, Low: 111, Close: 115, Volume: 1500, Timeframe: "1d"},
        {Time: time.Now(), Symbol: "TEST", Open: 115, High: 119, Low: 114, Close: 117, Volume: 1600, Timeframe: "1d"},
        {Time: time.Now(), Symbol: "TEST", Open: 117, High: 120, Low: 116, Close: 118, Volume: 1700, Timeframe: "1d"},
        {Time: time.Now(), Symbol: "TEST", Open: 118, High: 122, Low: 117, Close: 120, Volume: 1800, Timeframe: "1d"},
        {Time: time.Now(), Symbol: "TEST", Open: 120, High: 124, Low: 119, Close: 122, Volume: 1900, Timeframe: "1d"},
    }
    return data
}

func TestCalculateSMA(t *testing.T) {
    data := createTestOHLCVData()
    period := 5

    sma, err := CalculateSMA(data, period)
    if err != nil {
        t.Fatalf("Failed to calculate SMA: %v", err)
    }

    if len(sma) != len(data) {
        t.Errorf("Expected SMA length %d, got %d", len(data), len(sma))
    }

    // Check that first few values are 0 (not enough data)
    for i := 0; i < period-1; i++ {
        if sma[i] != 0 {
            t.Errorf("Expected SMA[%d] to be 0, got %f", i, sma[i])
        }
    }

    // Calculate expected SMA for the 5th element (index 4)
    expected := (102.0 + 106.0 + 108.0 + 110.0 + 113.0) / 5.0
    if sma[4] != expected {
        t.Errorf("Expected SMA[4] to be %f, got %f", expected, sma[4])
    }
}

func TestCalculateEMA(t *testing.T) {
    data := createTestOHLCVData()
    period := 5

    ema, err := CalculateEMA(data, period)
    if err != nil {
        t.Fatalf("Failed to calculate EMA: %v", err)
    }

    if len(ema) != len(data) {
        t.Errorf("Expected EMA length %d, got %d", len(data), len(ema))
    }

    // EMA should have values starting from the period-1 index
    if ema[period-1] == 0 {
        t.Errorf("Expected EMA[%d] to be non-zero", period-1)
    }
}

func TestCalculateRSI(t *testing.T) {
    data := createTestOHLCVData()
    period := 5

    rsi, err := CalculateRSI(data, period)
    if err != nil {
        t.Fatalf("Failed to calculate RSI: %v", err)
    }

    if len(rsi) != len(data) {
        t.Errorf("Expected RSI length %d, got %d", len(data), len(rsi))
    }

    // Check that RSI values are between 0 and 100
    for i := period; i < len(rsi); i++ {
        if rsi[i] < 0 || rsi[i] > 100 {
            t.Errorf("RSI[%d] = %f is out of range [0, 100]", i, rsi[i])
        }
    }
}

func TestCalculateMACD(t *testing.T) {
    data := createTestOHLCVData()
    fastPeriod := 3
    slowPeriod := 6
    signalPeriod := 3

    // Need more data for MACD
    for i := 0; i < 10; i++ {
        data = append(data, models.OHLCV{
            Time:      time.Now(),
            Symbol:    "TEST",
            Open:      120.0 + float64(i),
            High:      125.0 + float64(i),
            Low:       119.0 + float64(i),
            Close:     122.0 + float64(i),
            Volume:    2000 + int64(i*100),
            Timeframe: "1d",
        })
    }

    macd, err := CalculateMACD(data, fastPeriod, slowPeriod, signalPeriod)
    if err != nil {
        t.Fatalf("Failed to calculate MACD: %v", err)
    }

    if len(macd.MACD) != len(data) {
        t.Errorf("Expected MACD length %d, got %d", len(data), len(macd.MACD))
    }

    if len(macd.Signal) != len(data) {
        t.Errorf("Expected Signal length %d, got %d", len(data), len(macd.Signal))
    }

    if len(macd.Histogram) != len(data) {
        t.Errorf("Expected Histogram length %d, got %d", len(data), len(macd.Histogram))
    }
}

func TestCalculateBollingerBands(t *testing.T) {
    data := createTestOHLCVData()
    period := 5
    deviation := 2.0

    bb, err := CalculateBollingerBands(data, period, deviation)
    if err != nil {
        t.Fatalf("Failed to calculate Bollinger Bands: %v", err)
    }

    if len(bb.Upper) != len(data) {
        t.Errorf("Expected Upper band length %d, got %d", len(data), len(bb.Upper))
    }

    if len(bb.Middle) != len(data) {
        t.Errorf("Expected Middle band length %d, got %d", len(data), len(bb.Middle))
    }

    if len(bb.Lower) != len(data) {
        t.Errorf("Expected Lower band length %d, got %d", len(data), len(bb.Lower))
    }

    // Check that upper > middle > lower for valid data points
    for i := period - 1; i < len(data); i++ {
        if bb.Upper[i] <= bb.Middle[i] {
            t.Errorf("Upper band[%d] = %f should be greater than middle band = %f", i, bb.Upper[i], bb.Middle[i])
        }
        if bb.Middle[i] <= bb.Lower[i] {
            t.Errorf("Middle band[%d] = %f should be greater than lower band = %f", i, bb.Middle[i], bb.Lower[i])
        }
    }
}

func TestCalculateStochastic(t *testing.T) {
    data := createTestOHLCVData()
    kPeriod := 5
    dPeriod := 3

    stoch, err := CalculateStochastic(data, kPeriod, dPeriod)
    if err != nil {
        t.Fatalf("Failed to calculate Stochastic: %v", err)
    }

    if len(stoch.K) != len(data) {
        t.Errorf("Expected K length %d, got %d", len(data), len(stoch.K))
    }

    if len(stoch.D) != len(data) {
        t.Errorf("Expected D length %d, got %d", len(data), len(stoch.D))
    }

    // Check that K values are between 0 and 100
    for i := kPeriod - 1; i < len(stoch.K); i++ {
        if stoch.K[i] < 0 || stoch.K[i] > 100 {
            t.Errorf("K[%d] = %f is out of range [0, 100]", i, stoch.K[i])
        }
    }
}

func TestCalculateATR(t *testing.T) {
    data := createTestOHLCVData()
    period := 5

    atr, err := CalculateATR(data, period)
    if err != nil {
        t.Fatalf("Failed to calculate ATR: %v", err)
    }

    if len(atr) != len(data) {
        t.Errorf("Expected ATR length %d, got %d", len(data), len(atr))
    }

    // ATR values should be positive
    for i := period; i < len(atr); i++ {
        if atr[i] <= 0 {
            t.Errorf("ATR[%d] = %f should be positive", i, atr[i])
        }
    }
}

func TestIndicatorCalculator(t *testing.T) {
    data := createTestOHLCVData()
    calculator := NewIndicatorCalculator()

    // Test SMA calculation
    params := map[string]interface{}{
        "period": 5,
    }
    result, err := calculator.Calculate(data, "sma", params)
    if err != nil {
        t.Fatalf("Failed to calculate SMA through calculator: %v", err)
    }

    sma, ok := result.([]float64)
    if !ok {
        t.Fatalf("Expected []float64, got %T", result)
    }

    if len(sma) != len(data) {
        t.Errorf("Expected SMA length %d, got %d", len(data), len(sma))
    }

    // Test RSI calculation
    rsiResult, err := calculator.Calculate(data, "rsi", params)
    if err != nil {
        t.Fatalf("Failed to calculate RSI through calculator: %v", err)
    }

    rsi, ok := rsiResult.([]float64)
    if !ok {
        t.Fatalf("Expected []float64, got %T", rsiResult)
    }

    if len(rsi) != len(data) {
        t.Errorf("Expected RSI length %d, got %d", len(data), len(rsi))
    }

    // Test unknown indicator
    _, err = calculator.Calculate(data, "unknown", params)
    if err == nil {
        t.Errorf("Expected error for unknown indicator, got nil")
    }
}

func TestInsufficientData(t *testing.T) {
    // Test with insufficient data
    shortData := []models.OHLCV{
        {Time: time.Now(), Symbol: "TEST", Open: 100, High: 105, Low: 98, Close: 102, Volume: 1000, Timeframe: "1d"},
        {Time: time.Now(), Symbol: "TEST", Open: 102, High: 108, Low: 101, Close: 106, Volume: 1100, Timeframe: "1d"},
    }

    _, err := CalculateSMA(shortData, 5)
    if err == nil {
        t.Errorf("Expected error for insufficient data, got nil")
    }

    _, err = CalculateRSI(shortData, 5)
    if err == nil {
        t.Errorf("Expected error for insufficient data, got nil")
    }

    _, err = CalculateATR(shortData, 5)
    if err == nil {
        t.Errorf("Expected error for insufficient data, got nil")
    }
}

func BenchmarkCalculateSMA(b *testing.B) {
    data := createTestOHLCVData()
    
    // Extend data for better benchmarking
    for i := 0; i < 1000; i++ {
        data = append(data, models.OHLCV{
            Time:      time.Now(),
            Symbol:    "TEST",
            Open:      100.0 + float64(i%50),
            High:      105.0 + float64(i%50),
            Low:       95.0 + float64(i%50),
            Close:     102.0 + float64(i%50),
            Volume:    1000 + int64(i),
            Timeframe: "1d",
        })
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := CalculateSMA(data, 20)
        if err != nil {
            b.Fatalf("Failed to calculate SMA: %v", err)
        }
    }
}

func BenchmarkCalculateRSI(b *testing.B) {
    data := createTestOHLCVData()
    
    // Extend data for better benchmarking
    for i := 0; i < 1000; i++ {
        data = append(data, models.OHLCV{
            Time:      time.Now(),
            Symbol:    "TEST",
            Open:      100.0 + float64(i%50),
            High:      105.0 + float64(i%50),
            Low:       95.0 + float64(i%50),
            Close:     102.0 + float64(i%50),
            Volume:    1000 + int64(i),
            Timeframe: "1d",
        })
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := CalculateRSI(data, 14)
        if err != nil {
            b.Fatalf("Failed to calculate RSI: %v", err)
        }
    }
}
