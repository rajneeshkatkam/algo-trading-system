package indicators

import (
    "fmt"
    "math"

    "github.com/algo-trading/market-data-service/internal/models"
)

// MovingAverage types
type MAType int

const (
    SMA MAType = iota // Simple Moving Average
    EMA               // Exponential Moving Average
    WMA               // Weighted Moving Average
)

// CalculateSMA calculates Simple Moving Average
func CalculateSMA(data []models.OHLCV, period int) ([]float64, error) {
    if len(data) < period {
        return nil, fmt.Errorf("insufficient data: need %d, got %d", period, len(data))
    }

    result := make([]float64, len(data))
    
    for i := period - 1; i < len(data); i++ {
        sum := 0.0
        for j := i - period + 1; j <= i; j++ {
            sum += data[j].Close
        }
        result[i] = sum / float64(period)
    }

    return result, nil
}

// CalculateEMA calculates Exponential Moving Average
func CalculateEMA(data []models.OHLCV, period int) ([]float64, error) {
    if len(data) < period {
        return nil, fmt.Errorf("insufficient data: need %d, got %d", period, len(data))
    }

    result := make([]float64, len(data))
    multiplier := 2.0 / (float64(period) + 1.0)

    // Calculate initial SMA as first EMA value
    sum := 0.0
    for i := 0; i < period; i++ {
        sum += data[i].Close
    }
    result[period-1] = sum / float64(period)

    // Calculate EMA for remaining values
    for i := period; i < len(data); i++ {
        result[i] = (data[i].Close * multiplier) + (result[i-1] * (1 - multiplier))
    }

    return result, nil
}

// RSI represents Relative Strength Index
type RSI struct {
    Period int
    Values []float64
}

// CalculateRSI calculates Relative Strength Index
func CalculateRSI(data []models.OHLCV, period int) ([]float64, error) {
    if len(data) < period+1 {
        return nil, fmt.Errorf("insufficient data: need %d, got %d", period+1, len(data))
    }

    result := make([]float64, len(data))
    gains := make([]float64, len(data))
    losses := make([]float64, len(data))

    // Calculate price changes
    for i := 1; i < len(data); i++ {
        change := data[i].Close - data[i-1].Close
        if change > 0 {
            gains[i] = change
            losses[i] = 0
        } else {
            gains[i] = 0
            losses[i] = -change
        }
    }

    // Calculate initial average gain and loss
    avgGain := 0.0
    avgLoss := 0.0
    for i := 1; i <= period; i++ {
        avgGain += gains[i]
        avgLoss += losses[i]
    }
    avgGain /= float64(period)
    avgLoss /= float64(period)

    // Calculate RSI
    for i := period; i < len(data); i++ {
        if i == period {
            // First RSI calculation
            if avgLoss == 0 {
                result[i] = 100
            } else {
                rs := avgGain / avgLoss
                result[i] = 100 - (100 / (1 + rs))
            }
        } else {
            // Smooth the averages
            avgGain = ((avgGain * float64(period-1)) + gains[i]) / float64(period)
            avgLoss = ((avgLoss * float64(period-1)) + losses[i]) / float64(period)
            
            if avgLoss == 0 {
                result[i] = 100
            } else {
                rs := avgGain / avgLoss
                result[i] = 100 - (100 / (1 + rs))
            }
        }
    }

    return result, nil
}

// MACD represents Moving Average Convergence Divergence
type MACD struct {
    MACD      []float64
    Signal    []float64
    Histogram []float64
}

// CalculateMACD calculates MACD indicator
func CalculateMACD(data []models.OHLCV, fastPeriod, slowPeriod, signalPeriod int) (*MACD, error) {
    if len(data) < slowPeriod {
        return nil, fmt.Errorf("insufficient data: need %d, got %d", slowPeriod, len(data))
    }

    // Calculate fast and slow EMAs
    fastEMA, err := CalculateEMA(data, fastPeriod)
    if err != nil {
        return nil, fmt.Errorf("failed to calculate fast EMA: %w", err)
    }

    slowEMA, err := CalculateEMA(data, slowPeriod)
    if err != nil {
        return nil, fmt.Errorf("failed to calculate slow EMA: %w", err)
    }

    // Calculate MACD line
    macdLine := make([]float64, len(data))
    for i := slowPeriod - 1; i < len(data); i++ {
        macdLine[i] = fastEMA[i] - slowEMA[i]
    }

    // Create temporary OHLCV data for signal line calculation
    tempData := make([]models.OHLCV, len(data)-slowPeriod+1)
    for i := slowPeriod - 1; i < len(data); i++ {
        tempData[i-slowPeriod+1] = models.OHLCV{
            Close: macdLine[i],
        }
    }

    // Calculate signal line (EMA of MACD)
    signalEMA, err := CalculateEMA(tempData, signalPeriod)
    if err != nil {
        return nil, fmt.Errorf("failed to calculate signal EMA: %w", err)
    }

    // Map signal values back to original data length
    signalLine := make([]float64, len(data))
    for i := 0; i < len(signalEMA); i++ {
        signalLine[slowPeriod-1+i] = signalEMA[i]
    }

    // Calculate histogram
    histogram := make([]float64, len(data))
    for i := slowPeriod + signalPeriod - 2; i < len(data); i++ {
        histogram[i] = macdLine[i] - signalLine[i]
    }

    return &MACD{
        MACD:      macdLine,
        Signal:    signalLine,
        Histogram: histogram,
    }, nil
}

// BollingerBands represents Bollinger Bands
type BollingerBands struct {
    Upper  []float64
    Middle []float64
    Lower  []float64
}

// CalculateBollingerBands calculates Bollinger Bands
func CalculateBollingerBands(data []models.OHLCV, period int, deviation float64) (*BollingerBands, error) {
    if len(data) < period {
        return nil, fmt.Errorf("insufficient data: need %d, got %d", period, len(data))
    }

    // Calculate SMA (middle band)
    sma, err := CalculateSMA(data, period)
    if err != nil {
        return nil, fmt.Errorf("failed to calculate SMA: %w", err)
    }

    upper := make([]float64, len(data))
    lower := make([]float64, len(data))

    // Calculate standard deviation and bands
    for i := period - 1; i < len(data); i++ {
        // Calculate standard deviation
        sumSquares := 0.0
        for j := i - period + 1; j <= i; j++ {
            diff := data[j].Close - sma[i]
            sumSquares += diff * diff
        }
        stdDev := math.Sqrt(sumSquares / float64(period))

        upper[i] = sma[i] + (deviation * stdDev)
        lower[i] = sma[i] - (deviation * stdDev)
    }

    return &BollingerBands{
        Upper:  upper,
        Middle: sma,
        Lower:  lower,
    }, nil
}

// Stochastic represents Stochastic Oscillator
type Stochastic struct {
    K []float64 // %K line
    D []float64 // %D line
}

// CalculateStochastic calculates Stochastic Oscillator
func CalculateStochastic(data []models.OHLCV, kPeriod, dPeriod int) (*Stochastic, error) {
    if len(data) < kPeriod {
        return nil, fmt.Errorf("insufficient data: need %d, got %d", kPeriod, len(data))
    }

    k := make([]float64, len(data))

    // Calculate %K
    for i := kPeriod - 1; i < len(data); i++ {
        highest := data[i-kPeriod+1].High
        lowest := data[i-kPeriod+1].Low

        // Find highest high and lowest low in the period
        for j := i - kPeriod + 2; j <= i; j++ {
            if data[j].High > highest {
                highest = data[j].High
            }
            if data[j].Low < lowest {
                lowest = data[j].Low
            }
        }

        if highest == lowest {
            k[i] = 50 // Avoid division by zero
        } else {
            k[i] = ((data[i].Close - lowest) / (highest - lowest)) * 100
        }
    }

    // Calculate %D (SMA of %K)
    tempData := make([]models.OHLCV, len(data))
    for i := 0; i < len(data); i++ {
        tempData[i] = models.OHLCV{Close: k[i]}
    }

    d, err := CalculateSMA(tempData[kPeriod-1:], dPeriod)
    if err != nil {
        return nil, fmt.Errorf("failed to calculate %D: %w", err)
    }

    // Map %D back to original length
    dLine := make([]float64, len(data))
    for i := 0; i < len(d); i++ {
        dLine[kPeriod-1+i] = d[i]
    }

    return &Stochastic{
        K: k,
        D: dLine,
    }, nil
}

// ATR calculates Average True Range
func CalculateATR(data []models.OHLCV, period int) ([]float64, error) {
    if len(data) < period+1 {
        return nil, fmt.Errorf("insufficient data: need %d, got %d", period+1, len(data))
    }

    trueRanges := make([]float64, len(data))
    result := make([]float64, len(data))

    // Calculate True Range for each period
    for i := 1; i < len(data); i++ {
        tr1 := data[i].High - data[i].Low
        tr2 := math.Abs(data[i].High - data[i-1].Close)
        tr3 := math.Abs(data[i].Low - data[i-1].Close)
        
        trueRanges[i] = math.Max(tr1, math.Max(tr2, tr3))
    }

    // Calculate initial ATR (SMA of first 'period' true ranges)
    sum := 0.0
    for i := 1; i <= period; i++ {
        sum += trueRanges[i]
    }
    result[period] = sum / float64(period)

    // Calculate subsequent ATR values using smoothing
    for i := period + 1; i < len(data); i++ {
        result[i] = ((result[i-1] * float64(period-1)) + trueRanges[i]) / float64(period)
    }

    return result, nil
}

// IndicatorCalculator provides a unified interface for calculating indicators
type IndicatorCalculator struct{}

func NewIndicatorCalculator() *IndicatorCalculator {
    return &IndicatorCalculator{}
}

func (ic *IndicatorCalculator) Calculate(data []models.OHLCV, indicatorName string, params map[string]interface{}) (interface{}, error) {
    switch indicatorName {
    case "sma":
        period, ok := params["period"].(int)
        if !ok {
            return nil, fmt.Errorf("missing or invalid period parameter")
        }
        return CalculateSMA(data, period)

    case "ema":
        period, ok := params["period"].(int)
        if !ok {
            return nil, fmt.Errorf("missing or invalid period parameter")
        }
        return CalculateEMA(data, period)

    case "rsi":
        period, ok := params["period"].(int)
        if !ok {
            return nil, fmt.Errorf("missing or invalid period parameter")
        }
        return CalculateRSI(data, period)

    case "macd":
        fastPeriod, ok1 := params["fast_period"].(int)
        slowPeriod, ok2 := params["slow_period"].(int)
        signalPeriod, ok3 := params["signal_period"].(int)
        if !ok1 || !ok2 || !ok3 {
            return nil, fmt.Errorf("missing or invalid MACD parameters")
        }
        return CalculateMACD(data, fastPeriod, slowPeriod, signalPeriod)

    case "bollinger_bands":
        period, ok1 := params["period"].(int)
        deviation, ok2 := params["deviation"].(float64)
        if !ok1 || !ok2 {
            return nil, fmt.Errorf("missing or invalid Bollinger Bands parameters")
        }
        return CalculateBollingerBands(data, period, deviation)

    case "stochastic":
        kPeriod, ok1 := params["k_period"].(int)
        dPeriod, ok2 := params["d_period"].(int)
        if !ok1 || !ok2 {
            return nil, fmt.Errorf("missing or invalid Stochastic parameters")
        }
        return CalculateStochastic(data, kPeriod, dPeriod)

    case "atr":
        period, ok := params["period"].(int)
        if !ok {
            return nil, fmt.Errorf("missing or invalid period parameter")
        }
        return CalculateATR(data, period)

    default:
        return nil, fmt.Errorf("unknown indicator: %s", indicatorName)
    }
}
